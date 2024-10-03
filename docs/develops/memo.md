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

### hls
https://datatracker.ietf.org/doc/html/draft-pantos-http-live-streaming

## WebRTC
https://getstream.io/resources/projects/webrtc/advanced/bitrates-traffic/


### ffmpeg
https://github.com/leandromoreira/ffmpeg-libav-tutorial


ffmpeg -re -i ./test.mp4 -an -c:v libx264 -x264opts bframes=0 -x264-params keyint=30 -bsf:v h264_mp4toannexb -payload_type 127 -f rtp rtp://127.0.0.1:5000

ffmpeg -i input.mp4 -c copy -movflags frag_keyframe+empty_moov+default_base_moof output_fmp4.mp4

ffmpeg -i output_fmp4.mp4 -c:v copy -c:a libopus -b:a 128k output_opus.mp4

ffmpeg -i input.mp4 -c:v libx264 -profile:v baseline -level:v 3.1 -c:a copy output.mp4

## ingress
### files
curl -X POST http://127.0.0.1:8080/v1/ingress/files -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"path":"./test.mp4","mediaTypes":["video","audio"]}'
### rtp
curl -X POST http://127.0.0.1:8080/v1/ingress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":5000, "payloadType":127, "codecType":"h264"}'
ffmpeg -re -i ./test.mp4 -an -c:v libx264 -x264opts bframes=0 -x264-params keyint=30 -bsf:v h264_mp4toannexb -payload_type 127 -f rtp rtp://127.0.0.1:5000
curl -X POST http://127.0.0.1:8080/v1/ingress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":5003, "payloadType":96, "codecType":"opus"}'
ffmpeg -re -i ./test.webm -vn -c:a copy -payload_type 96 -f rtp rtp://127.0.0.1:5003

## egress
### rtp
curl -X POST http://127.0.0.1:8080/v1/egress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":6000, "mediaTypes":["audio"]}'
curl -X POST http://127.0.0.1:8080/v1/egress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":6000, "mediaTypes":["video"]}'
curl -X POST http://127.0.0.1:8080/v1/egress/rtp -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"addr":"127.0.0.1", "port":6000, "mediaTypes":["video","audio"]}'

### egress files
curl -X POST http://127.0.0.1:8080/v1/egress/files -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"path":"./output4","mediaTypes":["video","audio"],"interval":20000}'

### HLS
curl -X POST http://127.0.0.1:8080/v1/hls -H "Authorization: Bearer streamkey" -H "Content-Type: application/json" -d '{"path":"./output4","mediaTypes":["video","audio"],"interval":20000}'

pion
bluenviron
livekit


http://127.0.0.1:8080/v1/whip


## RTMP
https://veovera.org/docs/enhanced/enhanced-rtmp-v2.pdf
https://veovera.org/docs/legacy/amf0-file-format-spec.pdf


### jpeg
RGB -> YUV 로 변환 (ffmpeg에서는 jpeg encoder는 YUV를 사용)
YUV -> 4:2:0 으로 서브스케일링
8x8 픽셀 블록으로 분할 (별도 처리)
DCT (Discrete Cosine Transform) 수행 (저주파, 고주파등으로 변환)
양자화 (DCT 결과를 양자화하여 데이터를 압축, 고주파를 많이 압축함, 저주파는 덜 압축함. 일반적으로 고주파 성분에 덜 민감함)
런-길이 부호화 및 허프만 부호화
- run-length encoding: 같은 값이 반복되는 횟수를 압축 (연속된 0의 개수를 압축)
- huffman encoding: 빈도수가 높은 값에 짧은 코드를 할당하여 압축
  jpeg 생성.

손실 압축 방식. (손실 압축은 원본 이미지를 완벽하게 복원할 수 없음)

ffmpeg 에서 qscale (퀄리티 스케일) 과 compression_level (압축 레벨)을 조절하여 압축률을 조절할 수 있음.
qscale 이 1~31값을 지정 1이 최고 화질을 의미.  1에 가까울수록 주파수 성분들을 덜 양자화 함.
compression_level 이 0~9값을 지정 0이 최고 화질을 의미. run-length encoding과 huffman encoding에 영향. 값이 낮을수록 압축을 덜하고, 높을수록 강한 압축. 시간이 더 걸릴수 있다.


### png
RGB 그대로 사용.
필터링 사용. (이전 픽셀과의 차이를 사용하여 압축, 주파수가 적을수록 압축률이 크다)
deflate 알고리즘으로 압축 (LZ77 과 Huffman Coding) 을 함께 적용.
LZ77 반복된 문자부분을


### av1
https://aomediacodec.github.io/av1-isobmff/
### vp8
https://datatracker.ietf.org/doc/rfc6386/


bframes=0