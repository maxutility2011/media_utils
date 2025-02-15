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
	Channels string
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
		
		if strings.Contains(line, "#EXT-X-MAP") {
			leftQuote := strings.Index(line, "\"")
			rightQuote := strings.LastIndex(line, "\"")
			mapSegmentFileName := line[leftQuote+1 : rightQuote]
			mapSegmentUrl := varPlaylistBaseUrl + "/" + mapSegmentFileName
			fmt.Printf("Map segment URL: %s\n", mapSegmentUrl)

			_, _, err = downloadObject(mapSegmentUrl, dstFolder)
			if err != nil {
				fmt.Printf("Fail to download map segment: %s. Error: %v\n", mapSegmentUrl, err)
			}

			continue
		}

		if strings.Contains(prevLine, "#EXTINF") {
			segmentUrlStr := varPlaylistBaseUrl + "/" + line
			_, _, err = downloadObject(segmentUrlStr, dstFolder)
			if err != nil {
				fmt.Printf("Fail to download segment: %s. Error: %v\n", segmentUrlStr, err)
			}

			parts := strings.Split(line, "/")
			segFilename := parts[len(parts)-1]
			parts = parts[:len(parts)-1]

			var newFolder string = dstFolder
			for _, p := range parts {
				newFolder = newFolder + "/" + p
				
				_, err = os.Stat(newFolder)
				if errors.Is(err, os.ErrNotExist) {
					fmt.Printf("Path %s does not exist. Creating it...\n", newFolder)
					err = os.Mkdir(newFolder, 0777)
					if err != nil {
						fmt.Printf("Failed to mkdir: %s. Error: %v\n", newFolder, err)
						return err
					}
				}
			}

			src := dstFolder + "/" + segFilename // Assuming variant playlist and the segments are in the same folder.
			dest := newFolder + "/" + segFilename

			err = os.Rename(src, dest)
			fmt.Printf("Segment url: %s saved to: %s\n", segmentUrlStr, dest)
		}

		prevLine = line
	}

	return err
}

func parseMediaPlaylistData(mediaPlaylistUrl string, data []byte, dstFolder string) error {
	var err error
	if !downloadSegments {
		fmt.Println("Flag downloadSegments not set. Don't download segments\n")
		return nil
	}

	parsedURL, err := url.Parse(mediaPlaylistUrl)
	if err != nil {
		fmt.Printf("Error: Failed to parse variant HLS playlist URL: %s\n", mediaPlaylistUrl)
		return err
	}

	mediaPlaylistBaseUrl := parsedURL.Scheme + "://" + parsedURL.Host + filepath.Dir(parsedURL.Path)
	fmt.Printf("mediaPlaylistBaseUrl: %s\n", mediaPlaylistBaseUrl)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var prevLine string
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.Contains(line, "#EXT-X-MAP") {
			leftQuote := strings.Index(line, "\"")
			rightQuote := strings.LastIndex(line, "\"")
			mapSegmentFileName := line[leftQuote+1 : rightQuote]
			mapSegmentUrl := mediaPlaylistBaseUrl + "/" + mapSegmentFileName
			fmt.Printf("Map segment URL: %s\n", mapSegmentUrl)

			_, _, err = downloadObject(mapSegmentUrl, dstFolder)
			if err != nil {
				fmt.Printf("Fail to download map segment: %s. Error: %v\n", mapSegmentUrl, err)
			}

			continue
		}

		if strings.Contains(prevLine, "#EXTINF") {
			if strings.Contains(line, "#") {
				continue
			}

			segmentUrlStr := mediaPlaylistBaseUrl + "/" + line
			_, _, err = downloadObject(segmentUrlStr, dstFolder)
			if err != nil {
				fmt.Printf("Fail to download audio/subtitles segment: %s. Error: %v\n", segmentUrlStr, err)
			}

			parts := strings.Split(line, "/")
			segFilename := parts[len(parts)-1]
			parts = parts[:len(parts)-1]

			var newFolder string = dstFolder
			for _, p := range parts {
				newFolder = newFolder + "/" + p
				
				_, err = os.Stat(newFolder)
				if errors.Is(err, os.ErrNotExist) {
					fmt.Printf("Path %s does not exist. Creating it...\n", newFolder)
					err = os.Mkdir(newFolder, 0777)
					if err != nil {
						fmt.Printf("Failed to mkdir: %s. Error: %v\n", newFolder, err)
						return err
					}
				}
			}

			src := dstFolder + "/" + segFilename // Assuming variant playlist and the segments are in the same folder.
			dest := newFolder + "/" + segFilename

			err = os.Rename(src, dest)
			fmt.Printf("Segment url: %s saved to: %s\n", segmentUrlStr, dest)
		}

		prevLine = line
	}

	return err
}

func parseMediaInfo(line string) (string, Media) {
	var mediaTrack Media
	var mediaTrackId string

	posColon := strings.LastIndex(line, ":")
	prevEqual := posColon
	prevComma := posColon
	var key string
	var val string
	var i int
	for i=posColon; i<len(line); i++ {
		if line[i] == ',' {
			val = line[prevEqual+1 : i]

			if key == "TYPE" {
				mediaTrack.Type = val
				fmt.Printf("Media type: %s\n", val)
			} else if key == "GROUP-ID" {
				mediaTrack.Group_id = val
				fmt.Printf("Group_id: %s\n", val)
			} else if key == "NAME" {
				mediaTrack.Name = val
				fmt.Printf("Name: %s\n", val)
			} else if key == "LANGUAGE" {
				mediaTrack.Language = val
				fmt.Printf("Language: %s\n", val)
			} else if key == "AUTOSELECT" {
				mediaTrack.Auto_select = val
				fmt.Printf("Auto_select: %s\n", val)
			} else if key == "DEFAULT" {
				mediaTrack.Default = val
				fmt.Printf("Default: %s\n", val)
			} else if key == "CHANNELS" {
				mediaTrack.Channels = val
				fmt.Printf("Channels: %s\n", val)
			} else if key == "URI" {
				mediaTrack.Uri = val[1 : len(val)-1]
				fmt.Printf("Uri: %s\n", val)
			} else if key == "FORCED" {
				mediaTrack.Forced = val
				fmt.Printf("Forced: %s\n", val)
			} else if key == "INSTREAM-ID" {
				mediaTrack.Instream_id = val
				fmt.Printf("Instream_id: %s\n", val)
			}

			prevComma = i
		} else if line[i] == '=' {
			key = line[prevComma+1 : i]
			fmt.Printf("Key: %s\n", key)
			prevEqual = i
		}
	}

	val = line[prevEqual+1 : i]
	if key == "TYPE" {
		mediaTrack.Type = val
		fmt.Printf("Media type: %s\n", val)
	} else if key == "GROUP-ID" {
		mediaTrack.Group_id = val
		fmt.Printf("Group_id: %s\n", val)
	} else if key == "NAME" {
		mediaTrack.Name = val
		fmt.Printf("Name: %s\n", val)
	} else if key == "LANGUAGE" {
		mediaTrack.Language = val
		fmt.Printf("Language: %s\n", val)
	} else if key == "AUTOSELECT" {
		mediaTrack.Auto_select = val
		fmt.Printf("Auto_select: %s\n", val)
	} else if key == "DEFAULT" {
		mediaTrack.Default = val
		fmt.Printf("Default: %s\n", val)
	} else if key == "CHANNELS" {
		mediaTrack.Channels = val
		fmt.Printf("Channels: %s\n", val)
	} else if key == "URI" {
		mediaTrack.Uri = val[1 : len(val)-1]
		fmt.Printf("Uri: %s\n", val)
	} else if key == "FORCED" {
		mediaTrack.Forced = val
		fmt.Printf("Forced: %s\n", val)
	} else if key == "INSTREAM-ID" {
		mediaTrack.Instream_id = val
		fmt.Printf("Instream_id: %s\n", val)
	}
	
	//mediaTrack.Uri = val[1 : len(val)-1] // Assuming URI is always the last attribute.

	//fmt.Printf("mediaTrack.Uri: %s\n", mediaTrack.Uri)
	if mediaTrack.Uri != "" { 
		mediaTrackId = mediaTrack.Uri[ : strings.Index(mediaTrack.Uri, "/")]
		fmt.Printf("mediaTrackId: %s\n", mediaTrackId)
	}

	return mediaTrackId, mediaTrack
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
		}
	}

	renditionId = "video_" + rendition.Peak_bandwidth
	return renditionId, rendition
}

func parseMasterPlaylistData(data []byte) error {
	var err error
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var prevLine string
	var mediaSubfolder string
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.Contains(line, "#EXT-X-STREAM-INF") {
			rid, rend := parseRenditionInfo(line)
			renditionTable[rid] = rend
		}

		if strings.Contains(prevLine, "#EXT-X-STREAM-INF") {
			variantUrlStr := masterPlaylistBaseUrl + "/" + line
			mediaSubfolder = line[ : strings.Index(line, "/")]
			fmt.Printf("mediaSubfolder %s\n", mediaSubfolder)

			_, err = os.Stat(mediaSubfolder)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("Path %s does not exist. Creating it...\n", mediaSubfolder)
				err = os.Mkdir(mediaSubfolder, 0777)
				if err != nil {
					fmt.Println("Failed to mkdir: ", mediaSubfolder, " Error: ", err)
					return err
				}
			}

			var varData []byte
			var varPath string
			varData, varPath, err = downloadObject(variantUrlStr, mediaSubfolder)

			if err == nil {
				fmt.Printf("Variant playlist: %s downloaded to path: %s\n", variantUrlStr, varPath)
				parseVarPlaylistData(variantUrlStr, varData, mediaSubfolder)
			} else {
				fmt.Printf("Fail to download video variant playlist: %s. Error: %v\n", variantUrlStr, err)
			}
		}

		if strings.Contains(line, "#EXT-X-MEDIA") {
			mid, mediaTrack := parseMediaInfo(line)
			if mid == "" {
				continue
			}

			mediaTable[mid] = mediaTrack
			mediaSubfolder = mid

			if mediaTrack.Type == "AUDIO" {
				fmt.Printf("Audio track subfolder path: %s\n", mediaSubfolder)
			} else if  mediaTrack.Type == "SUBTITLES" {
				fmt.Printf("Subtitles track subfolder path: %s\n", mediaSubfolder)
			}

			mediaUrlStr := masterPlaylistBaseUrl + "/" + mediaTrack.Uri

			_, err = os.Stat(mediaSubfolder)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("Path %s does not exist. Creating it...\n", mediaSubfolder)
				err = os.Mkdir(mediaSubfolder, 0777)
				if err != nil {
					fmt.Println("Failed to mkdir: ", working_directory + "/" + mediaSubfolder, " Error: ", err)
					return err
				}
			}

			var mediaPlaylistData []byte
			var mediaPath string
			mediaPlaylistData, mediaPath, err = downloadObject(mediaUrlStr, mediaSubfolder)

			if err == nil {
				fmt.Printf("Media playlist: %s downloaded to path: %s\n", mediaUrlStr, mediaPath)
				parseMediaPlaylistData(mediaUrlStr, mediaPlaylistData, mediaSubfolder)
			} else {
				fmt.Printf("Fail to download audio/subtitles playlist: %s. Error: %v\n", mediaUrlStr, err)
			}
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
var mediaTable = make(map[string]Media)
var downloadSegments bool = false
var outputRenditionInfo bool = false
var keepSegmentStructure bool = true

func main() {
	playlistPtr := flag.String("playlist", "", "HLS playlist URL")
	wdPtr := flag.String("output", "", "Output folder (default to the working directory)")
	downloadSegmentsPtr := flag.String("downloadSegments", "0", "Whether or not to download segments, or parse the rendition info only")
	renditionInfoFlag := flag.String("renditionInfo", "1", "Whether or not to output rendition info")
	// If set to 0, media segments will be downloaded all under the rendition subfolder. Filename conflicts are possible.  
	keepSegmentStructureFlag := flag.String("keepSegmentStructure", "1", "Whether or not to keep the original segment structure")
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

	if *keepSegmentStructureFlag == "0" {
		keepSegmentStructure = false
	} else {
		keepSegmentStructure = true
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