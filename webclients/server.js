const express = require('express');
const http = require('http'); // HTTP 모듈 추가
const https = require('https');
const fs = require('fs');
const axios = require('axios');
const { createProxyMiddleware } = require('http-proxy-middleware');
const bodyParser = require('body-parser');
const path = require('path');

const privateKey = fs.readFileSync('./private.key', 'utf8');
const certificate = fs.readFileSync('./certificate.crt', 'utf8');
const credentials = { key: privateKey, cert: certificate };


const app = express();
app.use(bodyParser.json());

app.use(express.static(path.join(__dirname, 'public')));

app.post('/whip', async (req, res) => {
    const offer = req.body.offer;
    console.log('Received WHIP Offer:', offer);

    // TODO: 서버에서 WebRTC PeerConnection을 생성하고 Answer를 생성합니다.
    const answer = {}; // 생성된 Answer

    res.json({ answer });
});

app.post('/whep', async (req, res) => {
    res.sendFile(path.join(__dirname, 'public', 'webrtc/index.html'));
});


app.use("/*", createProxyMiddleware({
    target: `http://127.0.0.1:9090`,
    changeOrigin: true,
    pathRewrite: (path, req) => {
        console.log('rewrite path:', path, req.originalUrl);
        return req.originalUrl
    }
}))

// HTTP 서버 설정 (포트 3333)
const httpServer = http.createServer(app);
httpServer.listen(8080, () => {
    console.log('HTTP server is running on port 8080');
});

// HTTPS 서버 설정 (포트 3334)
const httpsServer = https.createServer(credentials, app);
httpsServer.listen(8081, () => {
    console.log('HTTPS server is running on port 8081');
});
