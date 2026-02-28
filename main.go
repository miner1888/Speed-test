package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var zeroChunk = make([]byte, 256*1024)

const htmlPage = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>⚡ 测速面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            background: linear-gradient(135deg, #0f0c29 0%, #302b63 50%, #24243e 100%); 
            color: #e0e0e0; 
            min-height: 100vh; 
            display: flex; 
            justify-content: center; 
            align-items: center; 
            padding: 20px;
        }
        .container { 
            background: rgba(255, 255, 255, 0.05); 
            backdrop-filter: blur(20px); 
            padding: 50px; 
            border-radius: 30px; 
            box-shadow: 0 25px 50px rgba(0,0,0,0.5), 0 0 0 1px rgba(255,255,255,0.1); 
            text-align: center; 
            width: 100%; 
            max-width: 480px; 
            border: 1px solid rgba(255,255,255,0.1);
            position: relative;
            overflow: hidden;
        }
        .container::before {
            content: '';
            position: absolute;
            top: -50%;
            left: -50%;
            width: 200%;
            height: 200%;
            background: radial-gradient(circle, rgba(99,102,241,0.1) 0%, transparent 70%);
            animation: pulse 4s ease-in-out infinite;
            pointer-events: none;
        }
        @keyframes pulse {
            0%, 100% { transform: scale(1); opacity: 0.5; }
            50% { transform: scale(1.1); opacity: 0.8; }
        }
        .ip-tag { 
            font-size: 13px; 
            color: #a5b4fc; 
            margin-bottom: 30px; 
            font-family: 'SF Mono', monospace; 
            background: rgba(99,102,241,0.1);
            padding: 8px 16px;
            border-radius: 20px;
            display: inline-block;
            border: 1px solid rgba(99,102,241,0.2);
        }
        .speed-box { 
            height: 140px; 
            display: flex; 
            flex-direction: column; 
            justify-content: center; 
            position: relative;
            z-index: 1;
        }
        .type-label { 
            font-size: 12px; 
            color: #6366f1; 
            text-transform: uppercase; 
            letter-spacing: 2px; 
            font-weight: 600;
            margin-bottom: 10px;
            text-shadow: 0 0 20px rgba(99,102,241,0.5);
        }
        .speed-value { 
            font-size: 72px; 
            font-weight: 800; 
            background: linear-gradient(135deg, #fff 0%, #a5b4fc 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            margin: 5px 0; 
            font-variant-numeric: tabular-nums; 
            line-height: 1;
        }
        .unit { 
            font-size: 20px; 
            color: #818cf8; 
            margin-left: 8px; 
            font-weight: 500;
        }
        .progress-container { 
            width: 100%; 
            height: 6px; 
            background: rgba(255,255,255,0.1); 
            border-radius: 3px; 
            margin: 35px 0; 
            overflow: hidden; 
            position: relative;
        }
        #progress-bar { 
            width: 0%; 
            height: 100%; 
            background: linear-gradient(90deg, #6366f1 0%, #8b5cf6 50%, #ec4899 100%); 
            transition: width 0.1s linear; 
            box-shadow: 0 0 20px rgba(99,102,241,0.5);
            border-radius: 3px;
        }
        button { 
            width: 100%; 
            background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%); 
            color: white; 
            border: none; 
            padding: 20px; 
            border-radius: 16px; 
            cursor: pointer; 
            font-size: 18px; 
            font-weight: 700; 
            transition: all 0.3s ease; 
            position: relative;
            overflow: hidden;
            text-transform: uppercase;
            letter-spacing: 1px;
            box-shadow: 0 10px 30px rgba(99,102,241,0.3);
        }
        button:hover:not(:disabled) { 
            transform: translateY(-2px); 
            box-shadow: 0 15px 40px rgba(99,102,241,0.4);
        }
        button:active:not(:disabled) {
            transform: translateY(0);
        }
        button:disabled { 
            background: rgba(255,255,255,0.1); 
            color: #6b7280; 
            cursor: not-allowed; 
            box-shadow: none;
        }
        .countdown {
            font-size: 48px;
            font-weight: 800;
            color: #ec4899;
            animation: bounce 1s ease infinite;
        }
        @keyframes bounce {
            0%, 100% { transform: scale(1); }
            50% { transform: scale(1.1); }
        }
        .results { 
            display: grid; 
            grid-template-columns: 1fr 1fr; 
            gap: 20px; 
            margin-top: 35px; 
            padding-top: 25px; 
            border-top: 1px solid rgba(255,255,255,0.1); 
        }
        .res-item { 
            font-size: 12px; 
            color: #9ca3af; 
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .res-val { 
            display: block; 
            font-size: 24px; 
            color: #fff; 
            margin-top: 8px; 
            font-weight: 700;
            background: linear-gradient(135deg, #fff 0%, #a5b4fc 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .cmd-section { 
            margin-top: 35px; 
            text-align: left; 
        }
        .cmd-label { 
            font-size: 11px; 
            color: #9ca3af; 
            margin-bottom: 10px; 
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .cmd-box { 
            background: rgba(0,0,0,0.3); 
            color: #a5b4fc; 
            padding: 14px; 
            border-radius: 12px; 
            font-family: 'SF Mono', Consolas, monospace; 
            font-size: 12px; 
            border: 1px solid rgba(99,102,241,0.2); 
            position: relative; 
            cursor: pointer; 
            margin-bottom: 8px;
            transition: all 0.3s ease;
        }
        .cmd-box:hover {
            border-color: rgba(99,102,241,0.5);
            background: rgba(0,0,0,0.4);
            transform: translateX(5px);
        }
        .cmd-desc { 
            font-size: 11px; 
            color: #6b7280; 
            margin-top: 6px; 
            margin-bottom: 20px;
            padding-left: 5px;
        }
        .divider {
            height: 1px;
            background: linear-gradient(90deg, transparent, rgba(99,102,241,0.3), transparent);
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="ip-tag">📍 你的IP: {{.IP}}</div>
        <div class="speed-box">
            <div id="type-label" class="type-label">准备就绪</div>
            <div class="speed-value"><span id="speed">0.00</span><span class="unit" id="unit">Mbps</span></div>
        </div>
        <div class="progress-container"><div id="progress-bar"></div></div>
        <button id="startBtn" onclick="runSpeedtest()">开始测速</button>
        <div class="results">
            <div class="res-item">下载速度<span class="res-val" id="dl-res">-</span></div>
            <div class="res-item">上传速度<span class="res-val" id="up-res">-</span></div>
        </div>
        <div class="cmd-section">
            <div class="cmd-label">🚀 Linux 命令行一键测速</div>
            <div class="cmd-box" onclick="copyCmd(this)">bash <(curl -sL http://{{.Addr}}/speed.sh)</div>
            
            <div class="cmd-label">📦 指定文件大小测速</div>
            <div class="cmd-box" onclick="copyCmd(this)">wget -O /dev/null http://{{.Addr}}/?size=500</div>
        </div>
    </div>
    <script>
        function format(bps) {
            if (bps < 1e6) return { v: (bps/1e3).toFixed(2), u: "Kbps" };
            if (bps < 1e9) return { v: (bps/1e6).toFixed(2), u: "Mbps" };
            return { v: (bps/1e9).toFixed(2), u: "Gbps" };
        }
        
        function sleep(ms) {
            return new Promise(resolve => setTimeout(resolve, ms));
        }
        
        async function runSpeedtest() {
            const btn = document.getElementById('startBtn');
            const typeLabel = document.getElementById('type-label');
            const speedEl = document.getElementById('speed');
            const unitEl = document.getElementById('unit');
            const progressBar = document.getElementById('progress-bar');
            
            btn.disabled = true;
            
            // 测试下载
            const dl = await testDownload(15);
            document.getElementById('dl-res').innerText = format(dl).v + " " + format(dl).u;
            
            // 间隔3秒倒计时
            typeLabel.innerText = "等待中";
            speedEl.innerHTML = '<span class="countdown">3</span>';
            unitEl.innerText = "";
            progressBar.style.width = "0%";
            
            for (let i = 3; i > 0; i--) {
                speedEl.innerHTML = '<span class="countdown">' + i + '</span>';
                await sleep(1000);
            }
            
            // 测试上传
            const up = await testUpload(15);
            document.getElementById('up-res').innerText = format(up).v + " " + format(up).u;
            
            typeLabel.innerText = "测速完成";
            speedEl.innerHTML = '0.00';
            unitEl.innerText = "Mbps";
            progressBar.style.width = "100%";
            btn.disabled = false;
        }
        
        function testDownload(dur) {
            return new Promise(resolve => {
                const start = performance.now();
                const end = start + dur * 1000;
                let bytes = 0;
                let lastUpdate = start;
                
                const ctrl = new AbortController();
                fetch('/download', { signal: ctrl.signal }).then(res => {
                    const reader = res.body.getReader();
                    
                    function push() {
                        reader.read().then(({done, value}) => {
                            const now = performance.now();
                            
                            if (done || now >= end) {
                                ctrl.abort();
                                resolve((bytes * 8) / dur);
                                return;
                            }
                            
                            bytes += value.length;
                            
                            // 每100ms更新一次UI，实现动态效果
                            if (now - lastUpdate > 100) {
                                const elapsed = (now - start) / 1000;
                                const speed = (bytes * 8) / elapsed;
                                const progress = Math.min((elapsed / dur) * 100, 100);
                                updateUI(speed, "正在测试下载...", progress);
                                lastUpdate = now;
                            }
                            
                            push();
                        });
                    }
                    push();
                });
            });
        }
        
        function testUpload(dur) {
            return new Promise(resolve => {
                const start = performance.now();
                const end = start + dur * 1000;
                let lastBytes = 0;
                let lastUpdate = start;
                
                // 创建一个大缓冲区用于上传
                const chunkSize = 1024 * 1024; // 1MB chunks
                const totalChunks = 2000; // 总共约2GB数据，足够15秒使用
                
                const xhr = new XMLHttpRequest();
                xhr.open("POST", "/upload");
                
                // 使用上传进度事件实现动态更新
                xhr.upload.onprogress = (e) => {
                    const now = performance.now();
                    
                    if (now >= end) {
                        xhr.abort();
                        const avgSpeed = (e.loaded * 8) / dur;
                        resolve(avgSpeed);
                        return;
                    }
                    
                    // 每100ms更新一次UI，与下载保持一致
                    if (now - lastUpdate > 100) {
                        const elapsed = (now - start) / 1000;
                        const currentSpeed = ((e.loaded - lastBytes) * 8) / ((now - lastUpdate) / 1000);
                        const progress = Math.min((elapsed / dur) * 100, 100);
                        
                        updateUI(currentSpeed, "正在测试上传...", progress);
                        
                        lastBytes = e.loaded;
                        lastUpdate = now;
                    }
                };
                
                xhr.onload = () => {
                    const now = performance.now();
                    const elapsed = Math.min((now - start) / 1000, dur);
                    resolve((lastBytes * 8) / elapsed);
                };
                
                xhr.onerror = () => {
                    resolve(0);
                };
                
                xhr.onabort = () => {
                    const now = performance.now();
                    const elapsed = Math.min((now - start) / 1000, dur);
                    resolve((lastBytes * 8) / elapsed);
                };
                
                // 发送数据
                const blob = new Blob([new Uint8Array(chunkSize)]);
                const formData = new FormData();
                
                // 使用 ReadableStream 来持续发送数据
                let sentChunks = 0;
                const stream = new ReadableStream({
                    pull(controller) {
                        if (sentChunks >= totalChunks || performance.now() >= end) {
                            controller.close();
                            return;
                        }
                        controller.enqueue(new Uint8Array(chunkSize));
                        sentChunks++;
                    }
                });
                
                // 使用 fetch 配合 ReadableStream 更好地控制上传
                fetch('/upload', {
                    method: 'POST',
                    body: stream,
                    duplex: 'half'
                }).catch(() => {});
                
                // 备用方案：使用传统方式
                const bigArray = new Uint8Array(100 * 1024 * 1024); // 100MB
                xhr.send(bigArray);
            });
        }
        
        function updateUI(bps, lab, prg) {
            const f = format(bps);
            document.getElementById('speed').innerText = f.v;
            document.getElementById('unit').innerText = f.u;
            document.getElementById('type-label').innerText = lab;
            document.getElementById('progress-bar').style.width = prg + "%";
        }
        
        function copyCmd(el) {
            navigator.clipboard.writeText(el.innerText);
            
            // 视觉反馈
            const original = el.innerText;
            el.innerText = "✓ 已复制";
            el.style.background = "rgba(99,102,241,0.3)";
            setTimeout(() => {
                el.innerText = original;
                el.style.background = "rgba(0,0,0,0.3)";
            }, 1500);
        }
    </script>
</head>
</html>
`

const shellScript = `#!/bin/bash
echo "🚀 测速地址: http://{{.Addr}}"
echo "--------------------------------------"

# 下载测试
echo -n "正在测试下载速度... "
DL_SPEED=$(curl -sL --max-time 15 -w "%{speed_download}" -o /dev/null http://{{.Addr}}/download 2>/dev/null)
DL_MBPS=$(awk -v s="$DL_SPEED" 'BEGIN {printf "%.2f", s * 8 / 1000000}')
echo "$DL_MBPS Mbps"

# 上传测试
echo -n "正在测试上传速度... "
UP_SPEED=$(dd if=/dev/zero bs=1M count=1000 2>/dev/null | timeout 15s curl -s -X POST -T - -w "%{speed_upload}" -o /dev/null http://{{.Addr}}/upload 2>/dev/null)
UP_MBPS=$(awk -v s="$UP_SPEED" 'BEGIN {printf "%.2f", s * 8 / 1000000}')
echo "$UP_MBPS Mbps"

echo "--------------------------------------"
echo "测速完成"
`

func main() {
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		sizeStr := r.URL.Query().Get("size")
		if sizeStr != "" {
			sizeMB, _ := strconv.Atoi(sizeStr)
			total := int64(sizeMB) * 1024 * 1024
			w.Header().Set("Content-Length", strconv.FormatInt(total, 10))
			for i := int64(0); i < total; i += int64(len(zeroChunk)) {
				w.Write(zeroChunk)
			}
			return
		}
		for {
			if _, err := w.Write(zeroChunk); err != nil {
				return
			}
		}
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/speed.sh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, strings.ReplaceAll(shellScript, "{{.Addr}}", r.Host))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sizeStr := r.URL.Query().Get("size")
		if sizeStr != "" {
			http.Redirect(w, r, "/download?size="+sizeStr, http.StatusTemporaryRedirect)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		content := strings.ReplaceAll(htmlPage, "{{.IP}}", getIP(r))
		content = strings.ReplaceAll(content, "{{.Addr}}", r.Host)
		fmt.Fprint(w, content)
	})

	fmt.Println("🔥 测速服务已启动，端口: 8760")
	http.ListenAndServe(":8760", nil)
}

func getIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}