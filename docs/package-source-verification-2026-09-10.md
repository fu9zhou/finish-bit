# 下载备用源核验（2026-09-10）

本次为 8 个原先单源的包补充备用地址；合并原有配置后，内嵌注册表的 11 个包、25 个平台产物均有备用源，3 个 OCR 附加资源也有备用源。新增 18 个不同的下载地址，全部与原注册表固定 SHA-256 一致。主源、版本、格式、解压路径、许可证和摘要均未改动。

## 来源与顺序

- GitHub 原发布地址保持第一优先级。
- ImageMagick 使用官方备用站，再使用 GH-Proxy。官方站身份依据 [ImageMagick 镜像说明](https://imagemagick.org/mirror/)；两个平台的具体固定版本文件均在官方 binaries 目录中。
- 7zip-bootstrap、7zip 使用项目的 SourceForge 分发地址；[7-Zip 官方下载页](https://www.7-zip.org/download.html)明确链接该分发渠道，[26.03 目录](https://sourceforge.net/projects/sevenzip/files/7-Zip/26.03/)列出对应文件。7zip-full 使用经过完整摘要校验的 GH-Proxy 备用地址。官网固定版本地址会重定向回 GitHub，不能作为独立备用链路；SourceForge 的 full 安装包候选未完成校验，二者均没有加入。
- pdfcpu、poppler、pandoc、tesseract 使用 GH-Proxy 作为备用传输。GH-Proxy 为第三方服务；不将它标为项目官方。
- OCR 模型和许可证使用 jsDelivr 的固定 Git 提交地址。jsDelivr 为第三方 CDN；提交和原始摘要保持固定。
- ffmpeg、ffprobe、qsv 的已有备用地址保持原样，本轮没有重新验证这些旧地址。

## 实现与回归

下载器补充了有限 HTTP 重试、本地 I/O 错误停止、调用方取消传播、失败文件清理及每个来源的错误汇总；OCR 附加资源复用同一路径。CLI/Web 展示来源序号、失败原因和切换目标。未增加用户自定义源配置。

回归中复现了 Windows 任务快照短暂共享锁导致读取失败的问题。新增真实文件锁测试先确认失败，再将快照读取限制为仅对 Windows 共享/锁冲突重试，最多等待 150 毫秒；持续锁定和其他错误仍返回。最小复现及原有跨进程/后台任务场景连续 5 轮通过。

最终 `go test ./...`、`go vet ./...`、`go build ./cmd/fnsh`、`node --test internal/web/ui_test.mjs`（13 项）均通过。新增下载回归覆盖 404、429、临时服务端错误、重试耗尽、长 Retry-After、传输中断、停滞、错误摘要、本地文件失败、取消、全部源失败及 OCR 附加资源换源安装。

## 验证方法与边界

对每个新增地址下载完整文件，计算 SHA-256 并与注册表比较。大文件在本地核验脚本中从同一来源续传；ImageMagick 官方站采用分段下载，逐段检查 Content-Range 和长度，拼接完整文件后校验。没有用主源或其他镜像的缓存补齐字节。

本次验证证明备用地址提供相同的固定文件，不代表所有平台都做过二进制运行验收，也不保证未来速度或可用性。产品仍顺序下载，换源重新开始；核验脚本的分段/续传不是新产品能力。

下载证据与辅助脚本保留在本地忽略目录 `.finishbit-test-sources/`；最终新增地址清单为 `registered.json`。下表为可随仓库保存的核验结果。

| 下载地址 | 完整字节数 | SHA-256 |
| --- | ---: | --- |
| [https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Windows_x86_64.zip](https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Windows_x86_64.zip) | 6225025 | `9809a70ee60ba78252628cc9738b284fbebf22bc2616ae903fb89b807e75a8a6` |
| [https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Linux_x86_64.tar.xz](https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Linux_x86_64.tar.xz) | 5603500 | `652830db95e81868dbe38fbb3f506365511c99a1831b5e8955deac38f3f645f8` |
| [https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Linux_arm64.tar.xz](https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Linux_arm64.tar.xz) | 4689192 | `41e452eb52ea735378b84d112aef47b27d7148ec02bdb4ac2de4eb7b6a8147cd` |
| [https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Darwin_x86_64.tar.xz](https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Darwin_x86_64.tar.xz) | 5635192 | `93fb1e782c8cad41b46f236812ff1570c6f026429564d5ede1295db427a3495c` |
| [https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Darwin_arm64.tar.xz](https://gh-proxy.com/https://github.com/pdfcpu/pdfcpu/releases/download/v0.15.0/pdfcpu_0.15.0_Darwin_arm64.tar.xz) | 4838192 | `2f49a15594dd46289d62219a1a2c7b26f0128ec0c658e655a2e62125db0c49ad` |
| [https://gh-proxy.com/https://github.com/oschwartz10612/poppler-windows/releases/download/v26.07.0-0/Release-26.07.0-0.zip](https://gh-proxy.com/https://github.com/oschwartz10612/poppler-windows/releases/download/v26.07.0-0/Release-26.07.0-0.zip) | 42202938 | `a711b0563b06edc488583d28198b6734c5a494afbbd1b9d87d3d2866062fb7e2` |
| [https://download.imagemagick.org/archive/binaries/ImageMagick-7.1.2-31-portable-Q16-x64.7z](https://download.imagemagick.org/archive/binaries/ImageMagick-7.1.2-31-portable-Q16-x64.7z) | 11739115 | `33d8b47bb404a6b30c672195795e64a8ce0c807f9bba084e359ff086d6c5d50d` |
| [https://gh-proxy.com/https://github.com/ImageMagick/ImageMagick/releases/download/7.1.2-31/ImageMagick-7.1.2-31-portable-Q16-x64.7z](https://gh-proxy.com/https://github.com/ImageMagick/ImageMagick/releases/download/7.1.2-31/ImageMagick-7.1.2-31-portable-Q16-x64.7z) | 11739115 | `33d8b47bb404a6b30c672195795e64a8ce0c807f9bba084e359ff086d6c5d50d` |
| [https://download.imagemagick.org/archive/binaries/ImageMagick-7.1.2-31-portable-Q16-arm64.7z](https://download.imagemagick.org/archive/binaries/ImageMagick-7.1.2-31-portable-Q16-arm64.7z) | 12220794 | `15809a3406fc22efe59338a5554bc3f56ca46bdb45e7157900d96e9416ec1fcd` |
| [https://gh-proxy.com/https://github.com/ImageMagick/ImageMagick/releases/download/7.1.2-31/ImageMagick-7.1.2-31-portable-Q16-arm64.7z](https://gh-proxy.com/https://github.com/ImageMagick/ImageMagick/releases/download/7.1.2-31/ImageMagick-7.1.2-31-portable-Q16-arm64.7z) | 12220794 | `15809a3406fc22efe59338a5554bc3f56ca46bdb45e7157900d96e9416ec1fcd` |
| [https://gh-proxy.com/https://github.com/jgm/pandoc/releases/download/3.11/pandoc-3.11-windows-x86_64.zip](https://gh-proxy.com/https://github.com/jgm/pandoc/releases/download/3.11/pandoc-3.11-windows-x86_64.zip) | 41761100 | `2ab72baf2399450e148ddf7a2a8689806c42e1bba71862b57e220fd9b8456d3d` |
| [https://downloads.sourceforge.net/project/sevenzip/7-Zip/26.03/7zr.exe](https://downloads.sourceforge.net/project/sevenzip/7-Zip/26.03/7zr.exe) | 602624 | `ad4c82fadcbdf93c03b4fc440f300509c7d60c5c2f4d183e35d9d70d6957037d` |
| [https://gh-proxy.com/https://github.com/ip7z/7zip/releases/download/26.03/7z2603-x64.exe](https://gh-proxy.com/https://github.com/ip7z/7zip/releases/download/26.03/7z2603-x64.exe) | 1661239 | `0859c524b8a63551848f0c246abddcb1d0b7b656b0fbfe879f8d85e61a9e6edd` |
| [https://gh-proxy.com/https://github.com/tesseract-ocr/tesseract/releases/download/5.5.3/tesseract-ocr-w64-setup-5.5.3.20260724.exe](https://gh-proxy.com/https://github.com/tesseract-ocr/tesseract/releases/download/5.5.3/tesseract-ocr-w64-setup-5.5.3.20260724.exe) | 26573224 | `bee9e3434bd94fd65387d9be28cd467a41f61b1275383b55b0f59a1331270ae4` |
| [https://cdn.jsdelivr.net/gh/tesseract-ocr/tessdata_fast@87416418657359cb625c412a48b6e1d6d41c29bd/chi_sim.traineddata](https://cdn.jsdelivr.net/gh/tesseract-ocr/tessdata_fast@87416418657359cb625c412a48b6e1d6d41c29bd/chi_sim.traineddata) | 2469156 | `a5fcb6f0db1e1d6d8522f39db4e848f05984669172e584e8d76b6b3141e1f730` |
| [https://cdn.jsdelivr.net/gh/tesseract-ocr/tessdata_fast@87416418657359cb625c412a48b6e1d6d41c29bd/eng.traineddata](https://cdn.jsdelivr.net/gh/tesseract-ocr/tessdata_fast@87416418657359cb625c412a48b6e1d6d41c29bd/eng.traineddata) | 4113088 | `7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2` |
| [https://cdn.jsdelivr.net/gh/tesseract-ocr/tessdata_fast@87416418657359cb625c412a48b6e1d6d41c29bd/LICENSE](https://cdn.jsdelivr.net/gh/tesseract-ocr/tessdata_fast@87416418657359cb625c412a48b6e1d6d41c29bd/LICENSE) | 11358 | `cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30` |
| [https://downloads.sourceforge.net/project/sevenzip/7-Zip/26.03/7z2603-extra.7z](https://downloads.sourceforge.net/project/sevenzip/7-Zip/26.03/7z2603-extra.7z) | 1764178 | `191894e6acb3647ffb69ce630479ff318523b2e2b9890aa7f05c1127c2e59b8f` |
