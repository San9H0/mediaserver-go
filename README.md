# mediaserver-go

## Overview
입력과 출력이 자유로운 미디어 스트리밍 서버

## Features
Live streams can be published to the server with:

| protocol         | variants  |video codecs|audio codecs|
|------------------|-----------|------------|------------|
| WebRTC Stream    | WHIP      | VP8, H264, AV1 | Opus |
| RTMP Stream      | RTMP      | H264 | AAC |
 | File Stream | mp4, webm | H264, VP8, AV1 | AAC, Opus |

And can be read from the server with:

| protocol      | variants  | video codecs   | audio codecs |
|---------------|-----------|----------------|--------------|
| WebRTC Client | WHEP      | VP8, H264, AV1 | Opus         |
| LL-HLS        | LL-HLS    | H264      | Opus, AAC    |
| Record File   | mp4, webm | H264, VP8, AV1 | AAC, Opus    |

## TODO
Adaptive Bitrate, Simulcast, SVC, RTMP AV1
