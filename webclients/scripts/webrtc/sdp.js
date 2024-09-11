
function parseSdpFmtpLine(sdpFmtpLine) {
    const paramsMap = {};
    if (sdpFmtpLine === undefined) {
        return paramsMap;
    }

    // ;를 기준으로 나누고 각 항목을 key-value 쌍으로 변환
    sdpFmtpLine.split(';').forEach(param => {
        const [key, value] = param.trim().split('=');
        paramsMap[key] = value;
    });

    return paramsMap;
}

