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

ffmpeg -i input.mp4 -c copy -movflags frag_keyframe+empty_moov+default_base_moof output_fmp4.mp4

ffmpeg -i output_fmp4.mp4 -c:v copy -c:a libopus -b:a 128k output_opus.mp4

ffmpeg -i input.mp4 -c:v libx264 -profile:v baseline -level:v 3.1 -c:a copy output.mp4

### curl
curl -X POST http://127.0.0.1:8080/v1/files -H "Authorization: Bearer fileKey" -H "Content-Type: application/json" -d '{"path":"./test.mp4","mediaType":"video"}'

## ingress
### rtp
curl -X POST http://127.0.0.1:8080/v1/ingress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":5000, "payloadType":127, "codecType":"h264"}'
ffmpeg -re -i ./test.mp4 -an -c:v libx264 -x264opts bframes=0 -x264-params keyint=30 -bsf:v h264_mp4toannexb -payload_type 127 -f rtp rtp://127.0.0.1:5000
curl -X POST http://127.0.0.1:8080/v1/ingress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":5003, "payloadType":96, "codecType":"opus"}'
ffmpeg -re -i ./test.webm -an -c:a copy -payload_type 96 -f rtp rtp://127.0.0.1:5003

## egress
### rtp
curl -X POST http://127.0.0.1:8080/v1/egress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":6000, "mediaTypes":["audio"]}'
curl -X POST http://127.0.0.1:8080/v1/egress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":6000, "mediaTypes":["video"]}'
curl -X POST http://127.0.0.1:8080/v1/egress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":6000, "mediaTypes":["video","audio"]}'

### egress files
curl -X POST http://127.0.0.1:8080/v1/egress/files -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"path":"./output4","mediaTypes":["video","audio"],"interval":20000}'

pion
bluenviron
livekit


http://127.0.0.1:8080/v1/whip