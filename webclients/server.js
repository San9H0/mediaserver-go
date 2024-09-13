const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const bodyParser = require('body-parser');
const path = require('path');


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
    target: `http://127.0.0.1:8080`,
    changeOrigin: true,
    pathRewrite: (path, req) => {
        return req.originalUrl
    }
}))

app.listen(3000, () => {
    console.log('Web server is running on port 3000');
});
