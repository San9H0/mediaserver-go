
function createReceiverPeerConnection() {
    // 새로운 RTCPeerConnection 생성
    peerConnection = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
    });

    // 원격 스트림을 수신할 때 처리
    peerConnection.ontrack = (event) => {
        remoteVideo.srcObject = event.streams[0];
    };

    const videoTx = peerConnection.addTransceiver('video', { direction: 'recvonly' });
    const audioTx = peerConnection.addTransceiver('audio', { direction: 'recvonly' });

    const codecs = RTCRtpSender.getCapabilities('video').codecs

    const h264FilteredCodec = codecs.filter(codec => {
        const sdpFmtpLines = parseSdpFmtpLine(codec.sdpFmtpLine)
        return codec.mimeType.toLowerCase() === 'video/h264' &&
            sdpFmtpLines['profile-level-id'] === '4d001f' &&
            sdpFmtpLines['packetization-mode'] === '1'})
    if (h264FilteredCodec.length === 0) {
        console.error('No suitable codec found')
        return
    }

    const opusFilteredCodec = codecs.filter(codec => {
        console.log("codec:", codec)
    })
    videoTx.setCodecPreferences((h264FilteredCodec))
}