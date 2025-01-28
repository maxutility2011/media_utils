A library of media utilities

**mp4_parser**
mp4_parser.go implements a mp4 box parser. It also offers method to retrieve "TFDT baseMediaDecodeTime" and "timescale" (Func GetTfdt) and set "TFDT baseMediaDecodeTime" and "timescale" (Func SetTfdtUint32). 

To build and run the test program: 
- cd test_mp4_parser
- go build test_mp4_parser_main.go
- ./test_mp4_parser_main -segment=2.mp4
