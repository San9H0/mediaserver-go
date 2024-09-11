# mediaserver

rtmp to hls

webrtc (whip, whep)

rtsp

srt


CGO_CFLAGS="-I/C/ffmpeg/include" CGO_LDFLAGS="-L/C/ffmpeg/lib -lavformat" go run .


curl -X POST http://127.0.0.1:8080/v1/files -H "Authorization: Bearer fileKey" -H "Content-Type: application/json" -d '{"path":"./test.mp4","mediaType":"video"}'


### avcc
https://stackoverflow.com/questions/24884827/possible-locations-for-sequence-picture-parameter-sets-for-h-264-stream

### avcc
https://heesu-choi.com/codec/h264-avcc-annexb/

### code
https://ffmpeg.org/doxygen/3.3/h264__mp4toannexb__bsf_8c_source.html
https://www.ffmpeg.org/doxygen/2.6/avc_8c_source.html#l00106


### ffmpeg
ffmpeg -re -i ./test.mp4 -an -c:v libx264 -x264opts bframes=0 -x264-params keyint=30 -bsf:v h264_mp4toannexb -payload_type 127 -f rtp rtp://127.0.0.1:5000