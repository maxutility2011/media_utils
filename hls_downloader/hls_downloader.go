package main

import (
	"fmt"
	"errors"
	"os"
	"strings"
	"flag"
	"bufio"
	"bytes"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"net/http"
)

func isValidHTTPURL(urlStr string) bool {
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

func downloadObject(urlStr string, dstFolder string) ([]byte, string, error) {
	var data []byte
	var local_path string
	var resp *http.Response
	var err error

	resp, err = http.Get(urlStr)
    if err != nil {
        fmt.Println("Error: Failed to download from: ", urlStr)
        return data, local_path, err
    }

    defer resp.Body.Close()

    data, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error: Failed to read response body")
        return data, local_path, err
    }

	var parsedURL *url.URL
	parsedURL, err = url.Parse(urlStr)
	if err != nil {
		fmt.Printf("Error: Failed to parse input URL: %s\n", urlStr)
		return data, local_path, err
	}
	
	local_path = working_directory + "/" + dstFolder + "/" + filepath.Base(parsedURL.Path)
	err = os.WriteFile(local_path, data, os.FileMode(0644)) // 0644: Read/write for owner, read for others

	if err != nil {
		fmt.Printf("Error: Failed to write to file: %s. Code: %v\n", local_path, err)
		return data, local_path, err
	}

    return data, local_path, err
}

func parseVarPlaylistData(varPlaylistUrl string, data []byte, dstFolder string) error {
	var err error

	parsedURL, err := url.Parse(varPlaylistUrl)
	if err != nil {
		fmt.Printf("Error: Failed to parse variant HLS playlist URL: %s\n", varPlaylistUrl)
		return err
	}

	varPlaylistBaseUrl := parsedURL.Scheme + "://" + parsedURL.Host + filepath.Dir(parsedURL.Path)
	fmt.Printf("varPlaylistBaseUrl: %s\n", varPlaylistBaseUrl)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var prevLine string
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.Contains(prevLine, "#EXTINF") {
			segmentUrlStr := varPlaylistBaseUrl + "/" + line
			var segPath string
			_, segPath, err = downloadObject(segmentUrlStr, dstFolder)

			fmt.Printf("Segment url: %s, path: %s\n", segmentUrlStr, segPath)
		}

		prevLine = line
	}

	return err
}

func parseMasterPlaylistData(data []byte) error {
	var err error
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var prevLine string
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.Contains(prevLine, "#EXT-X-STREAM-INF") {
			variantUrlStr := masterPlaylistBaseUrl + "/" + line

			posLastSlash := strings.LastIndex(line, "/")
			varPlaylistFilename := line[posLastSlash+1:] 
			posDot := strings.LastIndex(varPlaylistFilename, ".")
			varPlaylistFilenameNoExtension := varPlaylistFilename[:posDot]
			fmt.Printf("varPlaylistFilenameNoExtension: %s\n", varPlaylistFilenameNoExtension)

			variantSubfolder := working_directory + "/" + varPlaylistFilenameNoExtension + "/"
			_, err = os.Stat(variantSubfolder)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("Path %s does not exist. Creating it...\n", variantSubfolder)
				err = os.Mkdir(variantSubfolder, 0777)
				if err != nil {
					fmt.Println("Failed to mkdir: ", variantSubfolder, " Error: ", err)
					return err
				}
			}

			var varData []byte
			var varPath string
			varData, varPath, err = downloadObject(variantUrlStr, variantSubfolder)

			fmt.Printf("Variant playlist url: %s, path: %s\n", variantUrlStr, varPath)
			parseVarPlaylistData(variantUrlStr, varData, variantSubfolder)
		}

		prevLine = line
	}

	return err
}

var working_directory string
var input_playlist_local_path string
var isVariantPlaylist bool = false
var masterPlaylistBaseUrl string

func main() {
	playlistPtr := flag.String("playlist", "", "HLS playlist URL")
	wdPtr := flag.String("wd", "", "Working directory")
	flag.Parse()

	if *playlistPtr == "" {
		fmt.Printf("Input HLS playlist URL is required.\n")
		os.Exit(1)
	}

	var err error
	if *wdPtr != "" {
		working_directory = *wdPtr
		err = os.Chdir(*wdPtr)
		if err != nil {
			fmt.Printf("Error: Failed to change working directory to :%s\n", *wdPtr)
			os.Exit(1)
		}
	} else {
		working_directory = "."
	}

	if !isValidHTTPURL(*playlistPtr) {
		fmt.Printf("Invalid URL: %s\n", *playlistPtr)
		os.Exit(1)
	}

	parsedURL, err := url.Parse(*playlistPtr)
	if err != nil {
		fmt.Printf("Error: Failed to parse input HLS playlist URL: %s\n", *playlistPtr)
		os.Exit(1)
	}

	masterPlaylistBaseUrl = parsedURL.Scheme + "://" + parsedURL.Host + filepath.Dir(parsedURL.Path)
	fmt.Printf("masterPlaylistBaseUrl: %s\n", masterPlaylistBaseUrl)

	var downloadedData []byte
	var downloadedPath string
	downloadedData, downloadedPath, err = downloadObject(*playlistPtr, working_directory)
	if err != nil {
		fmt.Printf("Error: Failed to download playlist %s. Code: %v\n", *playlistPtr, err)
		os.Exit(1)
	}

	input_playlist_local_path = downloadedPath
	if !strings.Contains(string(downloadedData), "#EXT-X-STREAM-INF") {
		isVariantPlaylist = true
	} else {
		isVariantPlaylist = false
	}

	if !isVariantPlaylist {
		parseMasterPlaylistData(downloadedData)
	} else {
		parseVarPlaylistData(*playlistPtr, downloadedData, working_directory)
	}
}