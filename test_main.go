package main 

import (
	"fmt"
	"flag"
	"os"
	"io/ioutil"
	"media_utils/mutils"
)

func readSegment(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()
	segData, _ := ioutil.ReadAll(f)
	return segData, nil
}

func main() {
	segment_ptr := flag.String("segment", "", "file path")
	flag.Parse()

	seg_file_path := "segment.mp4"
	if *segment_ptr != "" {
		seg_file_path = *segment_ptr
	}

	seg_data, _ := readSegment(seg_file_path)
	fmt.Println("Read", len(seg_data), "bytes")

	//mutils.GetFtyp(seg_data)
	//var avc1 mutils.Avc1_box
	//avc1, _ = mutils.GetAvc1(seg_data)
	//fmt.Println("Video height:", avc1.Video_height, "Video width:", avc1.Video_width)

	var tfdt mutils.Tfdt_box
	tfdt, _ = mutils.GetTfdt(seg_data)
	fmt.Println("TFDT box size:", tfdt.Header.Box_size, "TFDT version:", tfdt.Header.Version, "BaseMediaDecodeTime V0:", tfdt.BaseMediaDecodeTime_v0, "BaseMediaDecodeTime V1:", tfdt.BaseMediaDecodeTime_v1)

	//mutils.getFtyp(seg_data)
	//mutils.getMoov(seg_data)
	//mutils.getMoof(seg_data)
	//mutils.getMdat(seg_data)
}