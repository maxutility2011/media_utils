package mutils

import (
	"fmt"
	"errors"
)

type Box_header struct {
	Box_size uint32
	Version uint8
	Flag uint32
}

type Avc1_box struct {
	Video_height uint16
	Video_width uint16
}

type Tfdt_box struct {
	Header Box_header
	BaseMediaDecodeTime_v0 uint32
	BaseMediaDecodeTime_v1 uint64
}

func get_uint8(p uint32, d []byte) uint8 {
	return d[p]
}

func get_uint16(p uint32, d []byte) uint16 {
	return uint16(d[p]) << 8 + uint16(d[p+1])
}

func get_uint32(p uint32, d []byte) uint32 {
	return uint32(d[p]) << 24 + uint32(d[p+1]) << 16 + uint32(d[p+2]) << 8 + uint32(d[p+3])
}

func get_uint64(p uint64, d []byte) uint64 {
	return uint64(d[p]) << 56 + uint64(d[p+1]) << 48 + uint64(d[p+2]) << 40 + uint64(d[p+3]) << 32 + uint64(d[p+4]) << 24 + uint64(d[p+5]) << 16 + uint64(d[p+6]) << 8 + uint64(d[p+7])
}

func mp4_fourcc(a byte, b byte, c byte, d byte) uint32 {
	return uint32(a) << 24 + uint32(b) << 16 + uint32(c) << 8 + uint32(d)
}

func GetFtyp(seg_data []byte) error {
	bytes_total := uint32(len(seg_data))
	bytes_remaining := bytes_total
	var box_size uint32
	var box_type uint32

	if bytes_remaining > 8 {
		box_size = get_uint32(0, seg_data)
		box_type = get_uint32(4, seg_data)
	} else {
		fmt.Println("Error: invalid segment data")
		return errors.New("Failed_to_find_ftyp")
	}

	for box_type != mp4_fourcc('f', 't', 'y', 'p') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("FTYP box not found")
			return errors.New("Failed_to_find_ftyp")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	//ftyp_start_offset := bytes_total - bytes_remaining
	ftyp_box_size := box_size
	fmt.Println("FTYP box size = ", ftyp_box_size)
	return nil
}

func GetMoof(seg_data []byte) error {
	bytes_total := uint32(len(seg_data))
	bytes_remaining := bytes_total
	var box_size uint32
	var box_type uint32

	if bytes_remaining > 8 {
		box_size = get_uint32(0, seg_data)
		box_type = get_uint32(4, seg_data)
	} else {
		fmt.Println("Error: invalid segment data")
		return errors.New("Failed_to_find_moof")
	}

	for box_type != mp4_fourcc('m', 'o', 'o', 'f') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("MOOF box not found")
			return errors.New("Failed_to_find_moof")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	//moof_start_offset := bytes_total - bytes_remaining
	moof_box_size := box_size
	fmt.Println("MOOF box size = ", moof_box_size)
	return nil
}

func GetMoov(seg_data []byte) error {
	bytes_total := uint32(len(seg_data))
	bytes_remaining := bytes_total
	var box_size uint32
	var box_type uint32

	if bytes_remaining > 8 {
		box_size = get_uint32(0, seg_data)
		box_type = get_uint32(4, seg_data)
	} else {
		fmt.Println("Error: invalid segment data")
		return errors.New("Failed_to_find_moov")
	}

	for box_type != mp4_fourcc('m', 'o', 'o', 'v') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("MOOV box not found")
			return errors.New("Failed_to_find_moov")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	//moov_start_offset := bytes_total - bytes_remaining
	moov_box_size := box_size
	fmt.Println("MOOV box size = ", moov_box_size)
	return nil
}

func GetMdat(seg_data []byte) error {
	bytes_total := uint32(len(seg_data))
	bytes_remaining := bytes_total
	var box_size uint32
	var box_type uint32

	if bytes_remaining > 8 {
		box_size = get_uint32(0, seg_data)
		box_type = get_uint32(4, seg_data)
	} else {
		fmt.Println("Error: invalid segment data")
		return errors.New("Failed_to_find_mdat")
	}

	for box_type != mp4_fourcc('m', 'd', 'a', 't') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("MDAT box not found")
			return errors.New("Failed_to_find_mdat")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	//mdat_start_offset := bytes_total - bytes_remaining
	mdat_box_size := box_size
	fmt.Println("MDAT box size = ", mdat_box_size)
	return nil
}

func GetTfdt(seg_data []byte) (Tfdt_box, error) {
	var tfdt Tfdt_box
	bytes_total := uint32(len(seg_data))
	bytes_remaining := bytes_total
	var box_size uint32
	var box_type uint32

	if bytes_remaining > 8 {
		box_size = get_uint32(0, seg_data)
		box_type = get_uint32(4, seg_data)
	} else {
		return tfdt, errors.New("Failed_to_find_tfdt")
	}

	for box_type != mp4_fourcc('m', 'o', 'o', 'f') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			return tfdt, errors.New("Failed_to_find_moof")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	moof_start_offset := bytes_total - bytes_remaining
	// skip MOOF payload
	box_size = get_uint32(moof_start_offset + 8, seg_data)
	box_type = get_uint32(moof_start_offset + 12, seg_data)
	bytes_remaining -= 8 

	for box_type != mp4_fourcc('t', 'r', 'a', 'f') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			return tfdt, errors.New("Failed_to_find_traf")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	traf_start_offset := bytes_total - bytes_remaining
	// skip TRAF payload
	box_size = get_uint32(traf_start_offset + 8, seg_data)
	box_type = get_uint32(traf_start_offset + 12, seg_data)
	bytes_remaining -= 8 

	for box_type != mp4_fourcc('t', 'f', 'd', 't') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			return tfdt, errors.New("Failed_to_find_tfdt")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	tfdt_start_offset := bytes_total - bytes_remaining
	tfdt_box_size := box_size

	tfdt.Header.Box_size = tfdt_box_size
	tfdt_version := get_uint8(tfdt_start_offset + 8, seg_data)
	if tfdt_version == 0 {
		tfdt.Header.Version = 0
		tfdt.BaseMediaDecodeTime_v0 = get_uint32(tfdt_start_offset + 12, seg_data)
	} else if tfdt_version == 1 {
		tfdt.Header.Version = 1
		tfdt.BaseMediaDecodeTime_v1 = get_uint64(uint64(tfdt_start_offset + 12), seg_data)
	}

	return tfdt, nil
}

func GetAvc1(seg_data []byte) (Avc1_box, error) {
	var avc1 Avc1_box
	bytes_total := uint32(len(seg_data))
	bytes_remaining := bytes_total
	var box_size uint32
	var box_type uint32

	if bytes_remaining > 8 {
		box_size = get_uint32(0, seg_data)
		box_type = get_uint32(4, seg_data)
	} else {
		fmt.Println("Error: invalid segment data")
		return avc1, errors.New("Failed_to_find_avc1")
	}

	for box_type != mp4_fourcc('m', 'o', 'o', 'v') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("MOOV box not found")
			return avc1, errors.New("Failed_to_find_moov")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	moov_start_offset := bytes_total - bytes_remaining
	// skip MOOV payload
	box_size = get_uint32(moov_start_offset + 8, seg_data)
	box_type = get_uint32(moov_start_offset + 12, seg_data)
	bytes_remaining -= 8 

	for box_type != mp4_fourcc('t', 'r', 'a', 'k') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("TRAK box not found")
			return avc1, errors.New("Failed_to_find_trak")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	// skip TRAK payload
	trak_start_offset := bytes_total - bytes_remaining
	box_size = get_uint32(trak_start_offset + 8, seg_data)
	box_type = get_uint32(trak_start_offset + 12, seg_data)
	bytes_remaining -= 8 

	for box_type != mp4_fourcc('m', 'd', 'i', 'a') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("MDIA box not found")
			return avc1, errors.New("Failed_to_find_mdia")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	// skip MDIA payload
	mdia_start_offset := bytes_total - bytes_remaining
	box_size = get_uint32(mdia_start_offset + 8, seg_data)
	box_type = get_uint32(mdia_start_offset + 12, seg_data)
	bytes_remaining -= 8 

	for box_type != mp4_fourcc('m', 'i', 'n', 'f') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("MINF box not found")
			return avc1, errors.New("Failed_to_find_minf")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	// skip MINF payload
	minf_start_offset := bytes_total - bytes_remaining
	box_size = get_uint32(minf_start_offset + 8, seg_data)
	box_type = get_uint32(minf_start_offset + 12, seg_data)
	bytes_remaining -= 8 

	for box_type != mp4_fourcc('s', 't', 'b', 'l') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("STBL box not found")
			return avc1, errors.New("Failed_to_find_stbl")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	// skip STBL payload
	stbl_start_offset := bytes_total - bytes_remaining
	box_size = get_uint32(stbl_start_offset + 8, seg_data)
	box_type = get_uint32(stbl_start_offset + 12, seg_data)
	bytes_remaining -= 8 

	for box_type != mp4_fourcc('s', 't', 's', 'd') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("STSD box not found")
			return avc1, errors.New("Failed_to_find_stsd")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	// skip STSD payload
	stsd_start_offset := bytes_total - bytes_remaining
	box_size = get_uint32(stsd_start_offset + 16, seg_data)
	box_type = get_uint32(stsd_start_offset + 20, seg_data)
	bytes_remaining -= 20

	for box_type != mp4_fourcc('a', 'v', 'c', '1') { 
		if (bytes_remaining < box_size + 8 || box_size == 0) {
			fmt.Println("AVC1 box not found")
			return avc1, errors.New("Failed_to_find_avc1")
		}

		bytes_remaining -= box_size
		box_size = get_uint32(bytes_total - bytes_remaining, seg_data)
		box_type = get_uint32(bytes_total - bytes_remaining + 4, seg_data)
	}

	avc1_start_offset := bytes_total - bytes_remaining
	avc1_box_size := box_size
	fmt.Println("AVC1 box size = ", avc1_box_size)

	avc1.Video_height = get_uint16(avc1_start_offset + 28, seg_data)
	avc1.Video_width = get_uint16(avc1_start_offset + 30, seg_data)
	return avc1, nil
}