A library of media utilities

**mp4_parser**
mp4_parser.go implements a mp4 box parser. It also offers method to retrieve "TFDT baseMediaDecodeTime" and "timescale" (Func GetTfdt) and set "TFDT baseMediaDecodeTime" and "timescale" (Func SetTfdtUint32). 

To build and run the test program: 
- cd test_mp4_parser
- go build test_mp4_parser_main.go
- ./test_mp4_parser_main -segment=2.mp4

**hls_downloader**
hls_downloader is a tool for downloading HLS playlists and media segments. 

./hls_downloader -downloadSegments=1 -keepSegmentStructure=1 -output=/Users/bo.zhang.-nd/Downloads/src/media_utils/hls_downloader/test/ -playlist=https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8 -renditionInfo=0
