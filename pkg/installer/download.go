package installer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func (i *Installer) installRepo(repo *Repo, tempDir string) error {
	var versionRegex = regexp.MustCompile(`_(v?\d+\.\d+\.\d+)`)
	archivePath, err := i.Download(repo.Links.ArchiveUrl, tempDir)
	if err != nil {
		sanitizedURL := versionRegex.ReplaceAllString(repo.Links.ArchiveUrl, "")
		archivePath, err = i.Download(sanitizedURL, tempDir)
		if err != nil {
			return fmt.Errorf("download failed with sanitized URL: %w", err)
		}
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	installedFiles, err := i.ExtractTarGz(file, tempDir)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	fmt.Printf("Successfully installed %s/%s. Files: %v\n", repo.Owner, repo.Name, installedFiles)
	return nil
}

func (i *Installer) Download(downloadURL, destDir string) (string, error) {
	fmt.Println("Starting download", "Download URL:", downloadURL)
	parsedURL, err := url.ParseRequestURI(downloadURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	fileName := filepath.Base(parsedURL.Path)
	if fileName == "" || fileName == "." {
		return "", fmt.Errorf("could not determine filename from URL")
	}

	fmt.Printf("Downloading %s...\n", fileName)

	filePath := filepath.Join(destDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	done := make(chan bool)
	go startSpinner(done)

	resp, err := i.client.Get(downloadURL)
	if err != nil {
		done <- true
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		done <- true
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		done <- true
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	done <- true

	fmt.Println("\nDownload complete.")
	return filePath, nil
}

func startSpinner(done chan bool) {
	spinner := []string{"|", "/", "-", "\\"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r")
			return
		default:
			fmt.Printf("\rDownloading... %s", spinner[i%len(spinner)])
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}

func (i *Installer) createTempDir() (string, error) {
	return os.MkdirTemp(i.config.TempDir, "install_*")
}

func (i *Installer) ExtractTarGz(gzipStream io.Reader, extractDir string) ([]string, error) {
	var installedFiles []string

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar reading error: %w", err)
		}

		// Construct the full path for the file or directory
		target := filepath.Join(extractDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it doesn't exist
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", target, err)
			}
		case tar.TypeReg:
			// Ensure the parent directory exists
			parentDir := filepath.Dir(target)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
			}

			// Extract the file
			if err := i.extractFile(tarReader, target, header.Mode); err != nil {
				return nil, err
			}

			// Handle executable files
			if header.Mode&0111 != 0 {
				destPath, err := i.moveToTargetDir(target)
				if err != nil {
					return nil, err
				}
				installedFiles = append(installedFiles, destPath)
			}
		}
	}

	return installedFiles, nil
}

func (i *Installer) extractFile(reader io.Reader, target string, mode int64) error {
	slog.Debug("Started Extraction")
	outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(mode))
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", target, err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to write file %s: %w", target, err)
	}

	return nil
}

func (i *Installer) moveToTargetDir(sourcePath string) (string, error) {
	if err := os.MkdirAll(i.config.TargetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	baseFile := filepath.Base(sourcePath)
	destFile := baseFile

	if strings.HasSuffix(baseFile, "_linux_amd64") {
		destFile = strings.TrimSuffix(baseFile, "_linux_amd64")
	}

	if strings.Contains(baseFile, "install-man-page") {
		slog.Info("Skipping man page installer script", "file", baseFile)
		return "", nil
	}

	destPath := filepath.Join(i.config.TargetDir, destFile)

	if err := copyFile(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy file to target directory: %w", err)
	}

	if err := os.Remove(sourcePath); err != nil {
		slog.Warn("failed to remove source file after copy", "error", err)
	}

	return destPath, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	if _, err := os.Stat(dst); err == nil {
		backupPath := dst + ".bak"
		if err := os.Rename(dst, backupPath); err != nil {
			slog.Warn("failed to create backup of existing file", "error", err)
		} else {
			slog.Info("Created backup of existing file", "backup", backupPath)
		}
	}

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
