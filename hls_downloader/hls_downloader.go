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
	"encoding/json"
)

type Rendition struct {
	Peak_bandwidth string
	Avg_bandwidth string
	Codecs string
	Resolution string
	Frame_rate string
	Hdcp_level string
	Video_range string
	Audio string
	Closed_captions string
}

type Media struct {
	Type string
	Uri string
	Group_id string
	Language string
	Name string
	Default string
	Auto_select string
	Forced string
	Instream_id string
}

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
	
	local_path = dstFolder + "/" + filepath.Base(parsedURL.Path)
	err = os.WriteFile(local_path, data, 0644) // 0644: Read/write for owner, read for others

	if err != nil {
		fmt.Printf("Error: Failed to write to file: %s. Code: %v\n", local_path, err)
		return data, local_path, err
	}

    return data, local_path, err
}

func parseVarPlaylistData(varPlaylistUrl string, data []byte, dstFolder string) error {
	var err error
	if !downloadSegments {
		fmt.Println("Flag downloadSegments not set. Don't download segments\n")
		return nil
	}

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

func parseRenditionInfo(line string) (string, Rendition) {
	var rendition Rendition
	var renditionId string

	posColon := strings.LastIndex(line, ":") + 1
	prevComma := posColon - 1
	prevEqual := posColon
	var codec1 string
	var codec2 string
	var key string
	var val string
	for i := posColon; i < len(line); i++ {
		if line[i] == ',' {
			if key != "CODECS" {
				val = line[prevEqual+1 : i]
				if key == "BANDWIDTH" {
					rendition.Peak_bandwidth = val
					fmt.Printf("Peak_bandwidth: %s\n", rendition.Peak_bandwidth)
				} else if key == "AVERAGE-BANDWIDTH" {
					rendition.Avg_bandwidth = val
					fmt.Printf("Avg_bandwidth: %s\n", rendition.Avg_bandwidth)
				} else if key == "RESOLUTION" {
					rendition.Resolution = val
					fmt.Printf("Resolution: %s\n", rendition.Resolution)
				} else if key == "FRAME-RATE" {
					rendition.Frame_rate = val
					fmt.Printf("Frame_rate: %s\n", rendition.Frame_rate)
				} else if key == "HDCP-LEVEL" {
					rendition.Hdcp_level = val
					fmt.Printf("Hdcp_level: %s\n", rendition.Hdcp_level)
				} else if key == "VIDEO-RANGE" {
					rendition.Video_range = val
					fmt.Printf("Video_range: %s\n", rendition.Video_range)
				} else if key == "AUDIO" {
					rendition.Audio = val
					fmt.Printf("Audio: %s\n", rendition.Audio)
				} else if key == "CLOSED-CAPTIONS" {
					rendition.Closed_captions = val
					fmt.Printf("Closed_captions: %s\n", rendition.Closed_captions)
				}
			} else {
				if codec1 == "" {
					codec1 = line[prevEqual+2 : i]
				} else {
					codec2 = line[prevComma : i-1]
					rendition.Codecs = codec1 + codec2
					codec1 = ""
					codec2 = ""
				}
			}

			prevComma = i
		} else if line[i] == '=' {
			prevEqual = i
			key = line[prevComma+1 : i]
			//fmt.Printf("Key: %s\n", key)
		}
	}

	renditionId = "video_" + rendition.Avg_bandwidth
	return renditionId, rendition
}

func parseMasterPlaylistData(data []byte) error {
	var err error
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var prevLine string
	var variantSubfolder string
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.Contains(line, "#EXT-X-STREAM-INF") {
			rid, rend := parseRenditionInfo(line)
			renditionTable[rid] = rend
			variantSubfolder = rid
			fmt.Printf("variantSubfolder: %s\n", variantSubfolder)
		}

		if strings.Contains(prevLine, "#EXT-X-STREAM-INF") {
			variantUrlStr := masterPlaylistBaseUrl + "/" + line

			/*
			posLastSlash := strings.LastIndex(line, "/")
			varPlaylistFilename := line[posLastSlash+1:] 
			posDot := strings.LastIndex(varPlaylistFilename, ".")
			varPlaylistFilenameNoExtension := varPlaylistFilename[:posDot]
			variantSubfolder := varPlaylistFilenameNoExtension + "/"
			*/

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

func dumpRenditionInfo() {
	rtBytes, err := json.MarshalIndent(renditionTable, "", "  ") // Pretty print
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile("renditions.json", rtBytes, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("JSON data written to renditions.json")
}

var working_directory string
var input_playlist_local_path string
var isVariantPlaylist bool = false
var masterPlaylistBaseUrl string
var renditionTable = make(map[string]Rendition)
var downloadSegments bool = false
var outputRenditionInfo bool = false

func main() {
	playlistPtr := flag.String("playlist", "", "HLS playlist URL")
	wdPtr := flag.String("output", "", "Output folder")
	downloadSegmentsPtr := flag.String("downloadSegments", "0", "Whether or not to download segments")
	renditionInfoFlag := flag.String("renditionInfo", "0", "Whether or not to output rendition info")
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

	if *downloadSegmentsPtr == "0" {
		downloadSegments = false
	} else {
		downloadSegments = true
	}

	if *renditionInfoFlag == "0" {
		outputRenditionInfo = false
	} else {
		outputRenditionInfo = true
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
		if outputRenditionInfo {
			dumpRenditionInfo()
		}
	} else {
		parseVarPlaylistData(*playlistPtr, downloadedData, working_directory)
	}
}