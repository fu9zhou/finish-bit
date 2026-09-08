# 帮小忙功能对齐核对清单（2026-09-07）

本轮只做批量核对和实施范围规划，未新增 Operation、安装依赖或提交 Git。对齐目标是可复用的本地能力；相同名称不意味着格式、质量、交互与运行环境完全相同。

## 核对基线与结论

- 网站：[首页](https://tool.browser.qq.com/)抓取 166 个唯一工具入口；9 个普通分类共 165 个唯一入口（165 次分类出现），首页额外出现“图片加水印”。浏览器插件分类另有 103 个扩展入口，见附录；首页学习工具横幅是外链推广，未算作工具。
- 项目：提交 `eb6e855e669294f1d7c77205e2541d4f5bf14872`，直接运行 `go run ./cmd/fnsh capabilities --json` 得到 170 个注册 Operation，并交叉读取 provider 与使用文档。未把站点工具数与 Operation 数作覆盖率除法。
- 方法：抓取公开首页与全部 10 个分类页；进一步读了 11 个容易误判的详情页。所有条目都有来源链接，但没有逐项登录、上传文件或实测站点服务。目录能力是初筛证据；未知的算法/格式矩阵在实施时补测，不称为完整行为验收。
- 已有表示主要离线功能和注册契约已存在；部分表示只有原语、子集或需要组合，不能标作完全对齐。待评估中同时包含重依赖与效果不确定，不能把它们一概当作做不了。

| 当前状态 | 网站入口数 |
| --- | ---: |
| 部分 | 22 |
| 待评估 | 47 |
| 环境依赖 | 5 |
| 数据待核 | 18 |
| 新增 | 51 |
| 已有 | 23 |

## 批次和轻量约束

| 批次 | 范围与完成标准 |
| --- | --- |
| R：复用 | 已有命令增加必要搜索别名/示例后做差异验收，不重复实现。 |
| A：优先实现 | 标准库或经过许可核实的小型 Go 库；换算、diff、编码、随机、二维码、文字处理。完整输入/输出与边界样例落入共享契约。 |
| B：随后实现 | 复用已托管 ImageMagick/pdfcpu/Poppler/Pandoc/FFmpeg，或小型生成算法；图像预设、PDF 组合、文档减重等。引擎已有不等于新包装已完成。 |
| D：数据核实后实施 | 生活/教育词库必须有可再分发许可、出处、版本和更新方式；使用用户自带数据可先交付引擎，但不算内置查询能力全完成。 |
| C：单独评估 | OCR、复杂 Office/PDF、AI、查杀等。先做一个可复现样例和依赖成本测量，不为了覆盖数量默认引入大型环境。 |
| E：环境/服务边界 | 浏览器捕获/编辑器交互、实时查询、搜索与账户服务；可抽出的本地算法进入 A/B，其余不伪装为 CLI 同等能力。 |

“库不能太大”尚未指定数字，建议作为后续实施默认预算而非用户既定限制：新增依赖带来的 stripped CLI 累计增量以 5 MiB 为评审线；单个新增离线数据包以 10 MiB 为评审线，超出则评估可选包。以同平台、同 Go 版本、同构建参数前后实测为准，不按仓库下载量猜二进制大小。所有候选必须同时核实传递依赖、嵌入词典、模型权重和字体许可。已有大引擎继续按需安装；不新增 Node/Python/Java 作为简单工具的常驻前置依赖。

候选库及官方来源见[轻量依赖研究](lightweight-tool-dependencies-2026-09-07.md)。纯公式、字符串转换优先标准库；二维码生成/扫描共用 gozxing；diff 优先自行实现，拼音、简繁和 YAML 按需各选一个库，具体版本、大小与接口在实施时固定和复核。

## 实施更新

2026-09-08：新增 107 个本地 Operation，合计 277 个。以下保留实施前基线；判断当前完成状态请看[逐项实施清单](qq-tool-implementation-2026-09-07.md)或 JSON 的 `implementation` 字段。OCR 中英文依赖约 78.4 MiB 已验收；LibreOffice 展开约 1.47 GiB，等待是否允许超过 1 GB 的例外。AI 权重许可与安装验证尚未闭环。

## 完整工具矩阵

“现有接口”是代码中已注册的 ID；“方案”中的新能力均为待实现。跨分类入口只计一次。

| # | 网站工具 / 来源 | 分类 | 状态 / 批次 | 现有接口 | 差距、实施方案与最小边界 |
| ---: | --- | --- | --- | --- | --- |
| 1 | [图片压缩](https://tool.browser.qq.com/tupianyasuo.html) | 图片 | 部分 / B | image.compress | 补目标字节数、批量压缩和结果是否变小的反馈；复用 ImageMagick，极限目标明确失败 |
| 2 | [证件照生成](https://tool.browser.qq.com/id_photo.html) | 图片 | 部分 / B/C | image.crop,image.resize,image.canvas | B 做标准尺寸、DPI、裁切和已透明背景换色；自动人像抠图需 C 模型，不能宣称全等价 |
| 3 | [PDF转Word](https://tool.browser.qq.com/pdf_2_word.html) | PDF | 部分 / B/C | pdf.extract-text,document.convert | B 组合文本提取到 DOCX（重排版）；扫描件 OCR 和复杂版式保真转 Word 归 C |
| 4 | [PDF转Excel](https://tool.browser.qq.com/pdf_2_excel.html) | PDF | 待评估 / C | pdf.text-boxes | 规则表格可按字坐标聚类再导出；合并单元格、无框表及扫描件需独立样本评测 |
| 5 | [PDF转HTML](https://tool.browser.qq.com/pdf_2_html.html) | PDF | 已有 / R | pdf.to-html | 已有 HTML 文件包导出；字体、复杂版面保真不等于网站效果已验收 |
| 6 | [PDF转图片](https://tool.browser.qq.com/pdf_2_png.html) | PDF | 部分 / B | pdf.render,image.join | 已有分页渲染，补一键长图合成及像素上限 |
| 7 | [PDF转PPT](https://tool.browser.qq.com/pdf_2_ppt.html) | PDF | 部分 / B | pdf.render,document.convert | 每页渲染图片后嵌入 PPTX；可展示但文字不可编辑，不算可编辑版式重建 |
| 8 | [Word转PDF](https://tool.browser.qq.com/word_2_pdf.html) | 文档 | 待评估 / C | — | Pandoc 当前无 PDF writer；完整 DOC/DOCX 排版需 LibreOffice 等大运行时，不纳入轻量核心 |
| 9 | [垃圾分类查询](https://tool.browser.qq.com/garbage.html) | 生活娱乐 | 数据待核 / D | — | 先核实可再分发分类数据及城市版本；小型离线字典与地区参数 |
| 10 | [Word格式转换](https://tool.browser.qq.com/word_convert.html) | 文档 | 部分 / B/C | document.convert | DOCX→HTML 已有；旧 DOC、PNG/PDF 需排版引擎，不能把文本转换算保真转换 |
| 11 | [PPT格式转换](https://tool.browser.qq.com/ppt_convert.html) | 文档 | 待评估 / C | — | 现有 Pandoc 不读 PPT/PPTX；图片、HTML、PDF 导出需新增 Office 渲染引擎 |
| 12 | [PDF加水印](https://tool.browser.qq.com/pdf_watermark.html) | PDF | 已有 / R | pdf.watermark,pdf.stamp | 已有文字、图片、PDF 水印和盖章；中文可通过图片资产处理 |
| 13 | [PDF瘦身](https://tool.browser.qq.com/pdf_compress.html) | PDF | 部分 / B | pdf.optimize,pdf.render,pdf.from-images | 结构优化已有；补扫描图降采样和目标大小，明确重栅格化会损失文本层 |
| 14 | [在线录屏](https://tool.browser.qq.com/screen_record.html) | 视频 | 环境依赖 / E | — | FFmpeg 现有输入仅本地文件；新增平台屏幕捕获或未来 Web getDisplayMedia，需交互选择与授权 |
| 15 | [Excel转PDF](https://tool.browser.qq.com/excel_to_pdf.html) | 文档 | 待评估 / C | — | XLS/XLSX 保留分页打印样式需 Office 引擎；简单数据表生成 PDF 只是子集 |
| 16 | [去手写](https://tool.browser.qq.com/handwriting_erasure.html) | 教育 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 17 | [文档瘦身](https://tool.browser.qq.com/office_reduce.html) | 文档 | 新增 / B | — | Go archive/zip + 图片重编码重打包 OOXML；先 DOCX/PPTX/XLSX，保留关系、宏和非图片项 |
| 18 | [身份证识别](https://tool.browser.qq.com/identification.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 19 | [银行卡识别](https://tool.browser.qq.com/bankcard_ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 20 | [印刷体识别](https://tool.browser.qq.com/basic_ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 21 | [手写体识别](https://tool.browser.qq.com/handwriting_ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 22 | [广告体识别](https://tool.browser.qq.com/ocr_advertise.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 23 | [英语识别](https://tool.browser.qq.com/ocr_english.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 24 | [护照识别](https://tool.browser.qq.com/passport_ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 25 | [PDF签名](https://tool.browser.qq.com/pdf_sign.html) | PDF | 部分 / B | pdf.stamp | 已有签名图片盖章；补坐标、大小、页选择；手绘属未来 UI；不宣称证书数字签名 |
| 26 | [运单识别](https://tool.browser.qq.com/waybill.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 27 | [集装箱识别](https://tool.browser.qq.com/container_ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 28 | [车辆证件识别](https://tool.browser.qq.com/carcard_ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 29 | [营业执照识别](https://tool.browser.qq.com/ocr_bizLicense.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 30 | [名片识别](https://tool.browser.qq.com/ocr_businesscard.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 31 | [港澳台证件识别](https://tool.browser.qq.com/ocr_permit.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 32 | [银行回单识别](https://tool.browser.qq.com/bankslip_ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 33 | [完税证明识别](https://tool.browser.qq.com/ocr_dutypaidproof.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 34 | [医疗票据识别](https://tool.browser.qq.com/ocr_recognize_medical_invoice.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 35 | [网约车行程单识别](https://tool.browser.qq.com/onlinetaxi_ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 36 | [人脸年龄选择](https://tool.browser.qq.com/face_age_transformation.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 37 | [图片编辑器](https://tool.browser.qq.com/img_edit_canvas.html) | 图片 | 部分 / B/E | image.crop,image.rotate,image.flip,image.annotate,image.adjust | CLI 编辑原语已有；补组合操作；画布、拖拽、撤销是 UI 层 |
| 38 | [密码安全检测](https://tool.browser.qq.com/password_check.html) | 开发 | 新增 / A | — | 本地长度、字符类别、重复模式和可选开源弱口令词表；结果是估计，不联网泄露密码 |
| 39 | [文件安全检测](https://tool.browser.qq.com/file_scan.html) | 开发 | 待评估 / C | file.checksum,file.info | 元数据和哈希不是恶意文件检测；ClamAV/病毒库很大，另设可选 provider 才可能覆盖 |
| 40 | [手机安装包检测](https://tool.browser.qq.com/app_inspector.html) | 开发 | 待评估 / C | archive.list | APK ZIP/manifest/权限静态检查可做子集；风险查杀、行为和信誉检测不等价 |
| 41 | [搜狗百宝箱](https://tool.browser.qq.com/sogou_box.html) | 生活娱乐 | 环境依赖 / E | — | 聚合跳转服务；不能当一个独立算法计入完成 |
| 42 | [合同验签](https://tool.browser.qq.com/contract_verification.html) | 生活娱乐 | 待评估 / C | pdf.validate | 结构校验不是合同验签；证书链、签署格式、时间戳与吊销状态需独立设计 |
| 43 | [合同对比](https://tool.browser.qq.com/contract_comparison.html) | 生活娱乐 | 部分 / B | document.text,pdf.extract-text | 提取文字后做段落 diff；扫描件和版式差异另评估，不能代替条款法律审查 |
| 44 | [图片转PDF](https://tool.browser.qq.com/img_2_pdf_convert.html) | PDF | 已有 / R | pdf.from-images | 有序图片生成 PDF；与站点格式矩阵逐项验收另做 |
| 45 | [PPT转PDF](https://tool.browser.qq.com/ppt_2_pdf.html) | 文档 | 待评估 / C | — | 同 PPT 转换，完整排版引擎超过轻量默认范围 |
| 46 | [搜狗PDF文件搜索](https://tool.browser.qq.com/pdf_search.html) | 生活娱乐 | 环境依赖 / E | — | 搜索索引属于线上服务；本地目录 PDF 搜索是另一个可新增的子集 |
| 47 | [PDF拆分](https://tool.browser.qq.com/pdf_split.html) | PDF | 已有 / R | pdf.split,pdf.select-pages | 已有按组拆分、选页输出 |
| 48 | [PDF合并](https://tool.browser.qq.com/pdf_merge.html) | PDF | 已有 / R | pdf.merge | 已有有序合并 |
| 49 | [转纯图PDF](https://tool.browser.qq.com/pdf_imagefy.html) | PDF | 部分 / B | pdf.render,pdf.from-images | 组合渲染→图片 PDF；补统一入口、分页尺寸及 DPI |
| 50 | [PDF页面管理](https://tool.browser.qq.com/pdf_page_manage.html) | PDF | 已有 / R | pdf.select-pages,pdf.remove-pages,pdf.insert-pages,pdf.rotate | 已有抽取、重排、删页、插空白页和旋转；可补从其他 PDF 插页 |
| 51 | [PDF图片提取](https://tool.browser.qq.com/pdf_img_extract.html) | PDF | 已有 / R | pdf.extract-images | 已有嵌入图片提取；不是语义图表识别 |
| 52 | [PDF加解密](https://tool.browser.qq.com/pdf_password.html) | PDF | 已有 / R | pdf.encrypt,pdf.decrypt | 已有 AES-256 及已知密码解密 |
| 53 | [PDF页面裁剪](https://tool.browser.qq.com/pdf_crop.html) | PDF | 部分 / B | pdf.crop | 已有等边距 CropBox；补四边独立边距/矩形坐标和逐页参数 |
| 54 | [修改PDF元数据](https://tool.browser.qq.com/pdf_metadata.html) | PDF | 部分 / B | pdf.info,pdf.keywords-add,pdf.keywords-remove | 仅关键词修改已有；补标题、作者、主题、创建信息等可写字段 |
| 55 | [字帖生成](https://tool.browser.qq.com/zitie_new.html) | 教育 | 新增 / B | — | Go SVG/HTML 模板生成田字格、米字格、描红；用户字体或开放字体，PDF 输出再复用已支持路径 |
| 56 | [Excel格式转换](https://tool.browser.qq.com/excel_convert.html) | 文档 | 待评估 / C | — | CSV/JSON 工具不能等于 XLS/XLSX→PNG/HTML/PDF；完整排版需要 Office 渲染 |
| 57 | [图片转Excel](https://tool.browser.qq.com/table_recognize.html) | 文档 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 58 | [今天吃什么](https://tool.browser.qq.com/whattoeat.html) | 生活娱乐 | 新增 / A | — | crypto/rand 从用户菜单抽取，可种子复现、排除项和批量抽签；默认菜单自建 |
| 59 | [文本转语音](https://tool.browser.qq.com/tts.html) | 文本 | 待评估 / C | — | 开源 TTS 引擎及声音数据按需安装；中文质量、模型许可与体积需要验证 |
| 60 | [提取文字](https://tool.browser.qq.com/ocr.html) | 图片 | 待评估 / C | — | 复用独立 OCR provider 后加此类字段/表格解析；先测 Tesseract+最小语言包，手写/特殊证件不保证达到网站效果；模型及语言包单算体积 |
| 61 | [亲戚关系计算](https://tool.browser.qq.com/relatives_name.html) | 教育 | 新增 / D | — | 亲属关系图规则、性别和地域称谓；规则集须自建或开源许可明确 |
| 62 | [PDF加页码](https://tool.browser.qq.com/pdf_page_number.html) | PDF | 部分 / B | pdf.stamp | 复用盖章实现起始页码、总页数、位置；核实动态占位符，不能把固定文本当完成 |
| 63 | [修改PDF页面尺寸](https://tool.browser.qq.com/pdf_pagesize.html) | PDF | 新增 / B | — | 扩展 pdfcpu provider 页面缩放/纸张规格，明确缩放内容还是仅改页面框 |
| 64 | [高校查询](https://tool.browser.qq.com/school.html) | 教育 | 数据待核 / D | — | 高校名称、地区、类型及版本化来源；招生等动态内容不能静态冒充最新 |
| 65 | [图片转字符画](https://tool.browser.qq.com/img_2_text.html) | 图片 | 新增 / B | — | 灰度采样映射字符；可选 ANSI/HTML 彩色输出，标准库即可 |
| 66 | [人脸性别转换](https://tool.browser.qq.com/face_gender_transformation.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 67 | [卡路里查询](https://tool.browser.qq.com/calories.html) | 生活娱乐 | 数据待核 / D | — | 食物热量查询与 calories_list/food_calories 共用一份有许可和计量单位的数据 |
| 68 | [增加图片大小](https://tool.browser.qq.com/img_enlarge.html) | 图片 | 部分 / B | image.resize | 尺寸放大已有；补最小文件字节数的编码策略，明确不增加真实细节 |
| 69 | [发票提取](https://tool.browser.qq.com/invoice_extract.html) | 图片 | 待评估 / C | — | OCR+发票字段/金额规范化；不同票据模板及扫描质量需评测 |
| 70 | [颜色转换](https://tool.browser.qq.com/colortrans.html) | 开发 | 新增 / A | — | RGB/HEX/HSL/HSV/CMYK 公式转换；可复用 go-colorful，说明 CMYK 无 ICC 色彩管理 |
| 71 | [图片日漫滤镜](https://tool.browser.qq.com/img_anime_filter.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 72 | [头像二次元](https://tool.browser.qq.com/img_face_anime.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 73 | [美化你的头像](https://tool.browser.qq.com/img_module_face_stylize.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 74 | [油画滤镜](https://tool.browser.qq.com/img_painting_filter.html) | 图片 | 新增 / B | — | 复用 ImageMagick 艺术滤镜；普通油画效果可达，网站模型风格不保证一致 |
| 75 | [人像分割](https://tool.browser.qq.com/img_human_split.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 76 | [照片美化](https://tool.browser.qq.com/img_pic_beauty.html) | 图片 | 部分 / B/C | image.adjust,image.sharpen,image.blur | 亮度饱和度锐化已有；补预设，磨皮和人脸语义美化归 C |
| 77 | [文档图片扭曲恢复](https://tool.browser.qq.com/doc_wrap.html) | 图片 | 待评估 / C | — | 四角透视校正可小算法做子集；弯曲纸张自动展平需视觉检测和评测 |
| 78 | [图片修复](https://tool.browser.qq.com/img_fix.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 79 | [图片去摩尔纹](https://tool.browser.qq.com/image_enhance.html) | 图片 | 待评估 / C | image.blur | 普通滤波不等于去摩尔纹；先评估频域方案，再决定是否需要模型 |
| 80 | [多人卡通画](https://tool.browser.qq.com/img_ai_cartoon.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 81 | [童话脸](https://tool.browser.qq.com/img_ai_face.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 82 | [乔丹风格化](https://tool.browser.qq.com/img_ai_jordanstyle.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 83 | [国漫水彩风](https://tool.browser.qq.com/img_ai_watercolor.html) | 图片 | 待评估 / C | — | 需要图像理解/分割/生成或修复模型；简单滤镜不等价，先核实引擎与权重许可、下载/内存及样本效果，不塞进核心 |
| 84 | [头像挂饰](https://tool.browser.qq.com/avatar_pendant.html) | 图片 | 部分 / B | image.watermark,image.canvas | 补开放或用户提供挂饰模板、裁切和批处理，不复制站点素材 |
| 85 | [字数计算](https://tool.browser.qq.com/wordcount.html) | 教育 | 已有 / R | text.count | 已有文本统计；中文计数口径与标点、空格定义需与样例对齐 |
| 86 | [英文创业公司/项目名生成](https://tool.browser.qq.com/startupname.html) | 生活娱乐 | 新增 / A | — | 用户词库/自建词根组合生成项目名；可设长度和数量，不承诺商标可用 |
| 87 | [手持弹幕LED](https://tool.browser.qq.com/led.html) | 生活娱乐 | 新增 / B/E | — | 生成本地 HTML/SVG 滚动弹幕或 GIF；全屏亮度与交互属浏览器 |
| 88 | [JSON diff](https://tool.browser.qq.com/jsondiff.html) | 开发 | 新增 / A | — | encoding/json + 结构化递归 diff；数组策略、路径、数值精度明确 |
| 89 | [JSON校验](https://tool.browser.qq.com/jsoncheck.html) | 开发 | 已有 / R | json.validate | 已有 JSON 校验 |
| 90 | [文本颜艺](https://tool.browser.qq.com/emoji.html) | 生活娱乐 | 新增 / A | — | 自建小型颜文字集合，筛选/随机输出 |
| 91 | [古诗词取名](https://tool.browser.qq.com/makename.html) | 生活娱乐 | 数据待核 / D | — | 公版诗词+署名和版本出处生成名字；选集整理数据也需许可检查 |
| 92 | [车牌归属地](https://tool.browser.qq.com/carnumber.html) | 生活娱乐 | 数据待核 / D | — | 离线省份/城市车牌前缀表，区分民用及特殊号牌，标版本 |
| 93 | [电话区号查询](https://tool.browser.qq.com/phonenumber.html) | 生活娱乐 | 数据待核 / D | — | 国内外电话区号字典；不是实时号码所有人或携号转网查询 |
| 94 | [英文人名生成](https://tool.browser.qq.com/fakeword.html) | 生活娱乐 | 新增 / A | — | 开源许可明确或自建名字词库随机组合，可复现 |
| 95 | [温度转换](https://tool.browser.qq.com/temperaturetrans.html) | 数据换算 | 新增 / A | — | 摄氏/华氏/开尔文，绝对零度与精度校验 |
| 96 | [投资收益计算](https://tool.browser.qq.com/invest.html) | 数据换算 | 新增 / A | — | 本金、利率、复利周期、追加投入计算；参数由用户给定，不提供投资预测 |
| 97 | [血型遗传规律](https://tool.browser.qq.com/bloodtype.html) | 生活娱乐 | 新增 / A | — | ABO 简化遗传枚举；清楚标模型假设，不用于亲子鉴定 |
| 98 | [五险一金计算](https://tool.browser.qq.com/wuxianyijin.html) | 数据换算 | 新增 / A/D | — | A 做用户输入比例/基数/封顶的计算；城市当期政策数据归 D |
| 99 | [md5加密](https://tool.browser.qq.com/md5.html) | 开发 | 已有 / R | hash.calculate | 支持 MD5 摘要；修正站点命名误导，MD5 不是加密或密码强度检测 |
| 100 | [进制转换](https://tool.browser.qq.com/hexconvert.html) | 数据换算 | 新增 / A | — | math/big 任意精度 2–36 进制整数，负数及非法位校验 |
| 101 | [简体繁体转换](https://tool.browser.qq.com/chinese.html) | 文本 | 新增 / A | — | 轻量开源简繁词典库；逐字/词组/地区变体及多义转换需要样例 |
| 102 | [房贷计算](https://tool.browser.qq.com/mortgage.html) | 数据换算 | 新增 / A | — | 等额本息、等额本金和还款计划；输入利率，不内置未经维护的现行政策 |
| 103 | [Unicode编解码](https://tool.browser.qq.com/unicode.html) | 开发 | 新增 / A | — | Unicode code point、JSON \u 转义、代理对与 UTF-8 校验；模式明确 |
| 104 | [url编解码](https://tool.browser.qq.com/urlencode.html) | 开发 | 已有 / R | url.encode,url.decode | 已有 query 编解码；路径段模式和表单 + 规则可扩展 |
| 105 | [数字大小写转换](https://tool.browser.qq.com/num2zh.html) | 数据换算 | 新增 / A | — | 中文数字及人民币大写，分角/负数/零与精度，用十进制定点处理 |
| 106 | [BMI计算](https://tool.browser.qq.com/bmi.html) | 数据换算 | 新增 / A | — | 体重身高换算并计算 BMI；只给数值和明确引用的分类标准 |
| 107 | [IP归属地](https://tool.browser.qq.com/iplocation.html) | 生活娱乐 | 数据待核 / D | — | IP 库许可、大小、更新和 IPv6 覆盖需核实；net/netip 只解析地址不能定位 |
| 108 | [文本比较](https://tool.browser.qq.com/textdiff.html) | 文本 | 新增 / A | — | 文本逐行/逐词 diff；Go 实现或 go-diff，明确大输入资源限制 |
| 109 | [JSON格式化](https://tool.browser.qq.com/jsonbeautify.html) | 开发 | 已有 / R | json.format,json.minify | 已有格式化和压缩 |
| 110 | [时间戳转换](https://tool.browser.qq.com/timestamp.html) | 数据换算 | 已有 / R | time.convert | 已有秒/毫秒/RFC3339/时区转换；补日期输入和边界测试可增强 |
| 111 | [在线加解密](https://tool.browser.qq.com/crypto.html) | 开发 | 新增 / A | — | 标准库 AES-GCM 文本/文件封装及必要兼容模式；先核实网站算法清单，不自行实现密码算法 |
| 112 | [GIF分解](https://tool.browser.qq.com/gifsplitter.html) | 图片 | 已有 / R | image.gif-split | 已有合成完整帧后输出 PNG |
| 113 | [URL解析](https://tool.browser.qq.com/urlparse.html) | 开发 | 新增 / A | — | net/url 解析 scheme/host/port/path/query/fragment，保留重复查询键 |
| 114 | [历史朝代查询](https://tool.browser.qq.com/dynasties.html) | 教育 | 数据待核 / D | — | 自建注明出处的历史时间表；并存朝代与公元前日期显式处理 |
| 115 | [各国首都](https://tool.browser.qq.com/capital.html) | 教育 | 数据待核 / D | — | 国家首都数据需要许可和更新日期；多首都/地区口径明确 |
| 116 | [日期计算](https://tool.browser.qq.com/datecal.html) | 数据换算 | 新增 / A | — | 日期加减/间隔/工作日；节假日表可由用户传入，时区和月末语义固定 |
| 117 | [计算器](https://tool.browser.qq.com/calculator.html) | 数据换算 | 新增 / A | — | 白名单表达式解析器，括号/幂/函数/精度及深度限制；不执行任意脚本 |
| 118 | [随机密码生成](https://tool.browser.qq.com/pwdgenerator.html) | 开发 | 新增 / A | — | crypto/rand、字符策略、长度、批量生成；不复用普通伪随机数生成口令 |
| 119 | [长度转换](https://tool.browser.qq.com/lengthconvert.html) | 数据换算 | 新增 / A | — | 长度单位表及精确比例；后续可扩面积/体积/质量 |
| 120 | [随机数生成](https://tool.browser.qq.com/random.html) | 生活娱乐 | 新增 / A | — | 区间整数、抽样/无放回、种子模式及密码学随机模式 |
| 121 | [base64编码](https://tool.browser.qq.com/base64.html) | 开发 | 已有 / R | base64.encode,base64.decode | 已有编码解码；核对 URL-safe、padding 等具体选项再增强 |
| 122 | [二维码生成](https://tool.browser.qq.com/qrcode.html) | 图片 | 新增 / A | — | gozxing 生成 PNG，内容/边距/尺寸/纠错级别 |
| 123 | [视频转gif](https://tool.browser.qq.com/video_2_gif.html) | 视频 | 已有 / R | video.gif | 已有 FFmpeg 视频转 GIF |
| 124 | [markdown编辑器](https://tool.browser.qq.com/markdown.html) | 开发 | 部分 / A/E | document.convert | 已有 Markdown→HTML 转换；goldmark 可补轻量预览输出，在线编辑器属 UI |
| 125 | [九宫格切图](https://tool.browser.qq.com/img9grid.html) | 图片 | 部分 / B | image.crop | 原语可裁切；补一键 NxM 分块、顺序清单和非整除边缘规则 |
| 126 | [图片格式转换](https://tool.browser.qq.com/imgconvert.html) | 图片 | 已有 / R | image.convert,image.formats | 已有 core PNG/JPEG 和可选 ImageMagick 格式；实际支持查运行时 |
| 127 | [GIF合成](https://tool.browser.qq.com/gifcreate.html) | 图片 | 已有 / R | image.gif-create | 已有帧序列合成 GIF |
| 128 | [文字转拼音](https://tool.browser.qq.com/tta.html) | 文本 | 新增 / A | — | go-pinyin 转拼音、声调/首字母/多音字选项；不承诺自动消歧 |
| 129 | [正则校验](https://tool.browser.qq.com/regexp.html) | 开发 | 部分 / A | text.replace | 原 text.replace 仅字面替换；新增正则校验、替换、匹配、分组和位置接口，明确 RE2 不支持回溯/环视 |
| 130 | [二维码美化](https://tool.browser.qq.com/prettify_qrcode.html) | 图片 | 新增 / B | — | 复用二维码+配色/图标叠加，输出前回扫校验；不破坏定位图案和静区 |
| 131 | [二维码名片](https://tool.browser.qq.com/visit_card.html) | 图片 | 新增 / A | — | vCard/MECARD 序列化+二维码，共用 QR 实现 |
| 132 | [快递查询](https://tool.browser.qq.com/kuaidi.html) | 生活娱乐 | 环境依赖 / E | — | 物流实时查询依赖承运商 API/凭据；单号格式识别不等于追踪 |
| 133 | [成语接龙](https://tool.browser.qq.com/jielong.html) | 教育 | 数据待核 / D | — | 成语库建立首尾字/拼音索引，与成语大全复用数据 |
| 134 | [医院推荐](https://tool.browser.qq.com/hospitalrecommend.html) | 生活娱乐 | 环境依赖 / E | — | 医院名录、地域、专科资料及更新来源；不以简单排序冒充医疗推荐 |
| 135 | [便捷思维导图](https://tool.browser.qq.com/markmap.html) | 教育 | 新增 / B | — | Markdown 标题解析生成可离线 HTML/SVG 树图；交互包单独资源，许可和体积待锁版 |
| 136 | [汉字标准发音](https://tool.browser.qq.com/hanzifayin.html) | 教育 | 待评估 / C/D | — | 拼音标注可复用 go-pinyin；标准读音播放需授权音频或 TTS，二者不等价 |
| 137 | [YAML/JSON互相转换](https://tool.browser.qq.com/yaml_2_json.html) | 开发 | 新增 / A | — | 维护中的 YAML 库；拒绝不兼容 JSON 的键，明确 alias/多文档/类型转换 |
| 138 | [图片黑白化](https://tool.browser.qq.com/img_fade.html) | 图片 | 已有 / R | image.grayscale | 已有灰度转换 |
| 139 | [元素周期表](https://tool.browser.qq.com/periodic.html) | 教育 | 数据待核 / D | — | 元素基本信息与来源、单位；小数据文件，资料许可需核实 |
| 140 | [食物热量表](https://tool.browser.qq.com/calories_list.html) | 生活娱乐 | 数据待核 / D | — | 与两处热量查询共用离线食物表，不重复开发三份 |
| 141 | [卡路里查询](https://tool.browser.qq.com/food_calories.html) | 生活娱乐 | 数据待核 / D | — | 与 calories/calories_list 共用查询、分量换算和来源 |
| 142 | [火星文翻译器](https://tool.browser.qq.com/toMars.html) | 文本 | 新增 / A | — | 自建可配置字符映射规则；明确非标准、不可逆，不冒充自然语言翻译 |
| 143 | [翻译](https://tool.browser.qq.com/translate.html) | 教育 | 待评估 / C/E | — | 本地翻译模型体积大；第三方 API 需要账户，词典替换不等价 |
| 144 | [汉字偏旁](https://tool.browser.qq.com/radical.html) | 教育 | 数据待核 / D | — | 汉字部首/笔画字典，明确字形和字典版本 |
| 145 | [歇后语](https://tool.browser.qq.com/allegory.html) | 教育 | 数据待核 / D | — | 歇后语问答查询与随机；整理词库许可先核实 |
| 146 | [词语注解](https://tool.browser.qq.com/explain.html) | 教育 | 数据待核 / D | — | 词语释义需要可再分发词典；不能复制站点释义库 |
| 147 | [世界时间校准](https://tool.browser.qq.com/timer.html) | 生活娱乐 | 部分 / A | time.convert | 补多城市当前时间输出；网页显示时间不等于修改操作系统时钟 |
| 148 | [成语大全](https://tool.browser.qq.com/chengyujielong.html) | 教育 | 数据待核 / D | — | 成语释义/搜索/接龙共用一库，避免与 jielong 重复 |
| 149 | [随机网名](https://tool.browser.qq.com/nick.html) | 生活娱乐 | 新增 / A | — | 自建前后缀/用户词库随机网名，可排除字符和重复 |
| 150 | [快递信息提取](https://tool.browser.qq.com/number_acquisition.html) | 生活娱乐 | 新增 / A/D | — | 正则提取姓名电话地址候选与单号，保留歧义；行政区解析要版本化数据 |
| 151 | [字节数换算](https://tool.browser.qq.com/byte_cal.html) | 数据换算 | 新增 / A | — | bit/byte、SI 1000 和 IEC 1024 分开，支持小数和精度 |
| 152 | [邮编查询](https://tool.browser.qq.com/zipcode.html) | 生活娱乐 | 数据待核 / D | — | 邮编地区映射可本地查，数据覆盖和许可需核实 |
| 153 | [userAgent工具](https://tool.browser.qq.com/useragent.html) | 开发 | 新增 / A | — | 解析浏览器/系统/设备及原始 UA；规则库若过大先做明确覆盖的子集 |
| 154 | [保质期计算](https://tool.browser.qq.com/shelflife.html) | 生活娱乐 | 新增 / A | — | 生产日期+天/月期限及到期剩余量；自然月和包含当天规则明确 |
| 155 | [取名字](https://tool.browser.qq.com/naming.html) | 生活娱乐 | 新增 / A/D | — | A 用户词库组合；D 开放中文名字素材，和其他取名入口共用生成器 |
| 156 | [文本去重](https://tool.browser.qq.com/unique.html) | 文本 | 已有 / R | text.unique | 已有按行去重；段落/大小写/保持顺序策略可扩展 |
| 157 | [图片像素化](https://tool.browser.qq.com/img_pixel.html) | 图片 | 新增 / B | — | Go 块平均/最近邻，或 ImageMagick 缩小放大；支持区域像素化 |
| 158 | [硬盘分区](https://tool.browser.qq.com/partition.html) | 数据换算 | 新增 / A | — | 仅实现 NTFS/FAT32 显示容量换算；详情使用说明是计算器，不操作磁盘 |
| 159 | [图片隐写](https://tool.browser.qq.com/image_secret_msg.html) | 图片 | 新增 / B | — | PNG 无损像素 LSB 编码/解码，自定义版本头、容量和校验；不是加密 |
| 160 | [开发语言输出hello world](https://tool.browser.qq.com/compilation.html) | 开发 | 新增 / A | — | 自建多语言 Hello World 模板，只返回源码，不安装或执行各语言运行时 |
| 161 | [文字隐写](https://tool.browser.qq.com/text_secret_msg.html) | 文本 | 新增 / A | — | 零宽字符编码/解码、版本和校验；提示文本规范化可能破坏载荷 |
| 162 | [uuid生成](https://tool.browser.qq.com/uuid.html) | 开发 | 已有 / R | uuid.generate | 已有 UUID 生成；不是加密功能 |
| 163 | [图片加水印](https://tool.browser.qq.com/watermark.html) | 图片（首页独有） | 已有 / R | image.watermark | 已有图片水印；文字可复用 annotate，批量水印预设可增强 |
| 164 | [表情包制作](https://tool.browser.qq.com/biaoqing.html) | 图片 | 部分 / B | image.annotate,image.gif-create | 补图片+文字模板、换行、字体、GIF 模板；不复制站点表情素材 |
| 165 | [二维码扫描](https://tool.browser.qq.com/qrcode_scan.html) | 图片 | 新增 / A | — | gozxing 从本地图片解码；摄像头扫描另由未来 UI 捕获图像，共用同一 Runner |
| 166 | [文档扫描](https://tool.browser.qq.com/word_scan.html) | 文档 | 待评估 / C | — | 扫描图增强、透视校正、OCR→DOCX 组合；当前无 OCR，先验证轻量候选 |

## 容易高估覆盖的项目

1. Pandoc 当前不读 PPTX、不支持 PDF 输出；现有 CSV/JSON 工具也不等于 Excel 排版渲染。[项目文档](../documents.md)
2. PDF→Word 的纯文本重排、PDF→PPT 的整页图片、通用 OCR→证件字段提取都需要标明输出层次，不能宣传为可编辑版式保真。[PDF 契约](../pdf.md)
3. 网站 PDF 签名使用说明出现手绘笔刷，因此签名图盖章是合理最小对齐；这不意味着我们有证书签名或法律效力验证。[详情](https://tool.browser.qq.com/pdf_sign.html)
4. 网站“硬盘分区”使用说明是输入容量并计算 NTFS/FAT32，纳入纯换算，不触碰磁盘分区表。[详情](https://tool.browser.qq.com/partition.html)
5. 网站部分营销说明有误：MD5 是摘要、UUID 是标识符，不能沿用其加密/安全宣传；部分 PPT/Word 页面说明互相混写，格式列表需真实输入验证。
6. 普通图像缩放、调色、模糊不能等于超分、去摩尔纹、背景分割或图像修复；可先交付确定性的子能力，但维持“部分”状态。[图片契约](../image-enhancements.md)

## 实施顺序与验收

1. A1：进制/单位/金额/日期/URL/Unicode/随机/口令等标准库工具；A2：JSON/text diff、正则独立接口、YAML、拼音简繁、二维码；A3：用户参数驱动的计算器、名字/菜单生成。每组交付后更新本矩阵。
2. B1：图片九宫格/像素/字符画/预设/二维码美化；B2：PDF 纯图、页码、签名、元数据、文字转 DOCX/图片转 PPTX；B3：OOXML 减重、字帖和离线思维导图。重复入口复用同一个引擎。
3. D 的数据授权核验与 A/B 开发可独立推进；C/E 先记录可交付子集、效果与依赖成本，不用半成品给矩阵打勾。
4. 复用用例放 `pkg/app`、契约放 `pkg/operation`、实现经 `operation.Runner`；CLI 保持薄层，脚本仅作 provider/开发验收，不绕过共享契约。
5. 每个新增或增强 Operation 都验证 direct CLI 与结构化调用、资源边界、错误码和取消/输出保留；涉及文件转换验证产物内容/尺寸/页数。代码变更完成前运行 `go test ./...`、`go vet ./...`、`go build ./cmd/fnsh`，并按现有约定更新生成目录和文档。

## 浏览器插件附录（103 个入口，单独计数）

来源：[浏览器插件分类](https://tool.browser.qq.com/category/pc_plugin)。这是扩展分发目录，存在同名不同 ID 和对外部产品的集成。下表只核对入口及可抽离方向，不证明插件本体开源、仍可用或可直接引入；未安装扩展，也未下载 CRX。浏览器侧效果不能用文件处理能力代替。

| # | 插件入口 | 扩展 ID | 对齐处理 |
| ---: | --- | --- | --- |
| 1 | Tampermonkey | dhdgffkkebhmkfjojejmpbldmpobfkfo | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 2 | 五彩网址二维码 | kojjendakpnlcgepocgjlmihheljihaj | 二维码生成/识别共用 A/B；当前网页读取和浮层由浏览器适配 |
| 3 | 有道词典Chrome划词插件 | eopjamdnofihpioajgfdikhhbobonhbb | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 4 | 小丰二维码 | anamdmjnllfgnoamcnlafmhemfcppbbc | 二维码生成/识别共用 A/B；当前网页读取和浮层由浏览器适配 |
| 5 | 常用工具(ToolBox) | fmmbocgmijhikfppllmnamafcphnelgn | JSON、编码、计算、二维码等按主清单复用；浏览器插件界面不是现有能力 |
| 6 | 橙光日历 | ocbgooipmfjolciehamjinlnjegkhlko | 日期/农历与时间规则可拆轻量能力；提醒、课程表和实时假日数据另行设计 |
| 7 | ACG助手 - 提供哔哩哔哩(bilibili)视频下载消息推送 | kpbnombpnpcffllnianjibmpadjolanh | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 8 | ICBCNewChromeExtension | ajmecfihhnibjmmihpecefjjckgbmedh | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 9 | 电脑管家上网防护 | ibgigpdnkkdnicediiebbfnednhmlpab | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 10 | 蓝岚日历 | ldabfnoapcgloeobngjnhckcolcfclli | 日期/农历与时间规则可拆轻量能力；提醒、课程表和实时假日数据另行设计 |
| 11 | Foxit PDF Creator | cifnddnffldieaamihfkhkdgnbhfmaci | PDF 文件渲染已有；浏览器内查看、网页打印需独立适配 |
| 12 | 替换字体的中文部分为雅黑 | enpkigfhoabjjjonanmddidnnahopmcn | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 13 | 打骨折-淘宝优惠券 天猫优惠券领取 | bodhjcioepfmdioecpjonecadajiiamj | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 14 | 小说阅读助手 | dknlfmhongfkfakmhhnmgfgnhhcbmldm | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 15 | NoteExpress网络捕手 | ljbhddngkkppbbkknjldoikonnolgafd | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 16 | 关灯看视频 | bfbmjmiodbnnpllbbbfblcplfjjepjdn | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 17 | 魔方换算 | kccajphaffpdfjhgcbccfmcpbgogbknh | JSON、编码、计算、二维码等按主清单复用；浏览器插件界面不是现有能力 |
| 18 | 健康提醒 | gefpceefdmmojbgmfmnkeanmpclobjgj | 日期/农历与时间规则可拆轻量能力；提醒、课程表和实时假日数据另行设计 |
| 19 | Downloads | jfchnphgogjhineanplmfkofljiagjfb | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 20 | IDM Integration Module | ngpampappnmepgilojfohadhhmbhlaek | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 21 | EditThisCookie | fngmhnnpilhplaeedifhccceomclgfbg | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 22 | 高校课程表 | diamegoomkocpgkfkmlmegiegkmogpao | 日期/农历与时间规则可拆轻量能力；提醒、课程表和实时假日数据另行设计 |
| 23 | ICBC Chrome Extension from Tendyron | dlombpffcodogboaljnamhpphpdkjdam | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 24 | Skype | lifbcibllhkdhoafpjfnlhfpfgnpldfl | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 25 | 划词翻译 | ikhdkkncnoglghljlkmcimlnlhkeamad | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 26 | 下载+ | gokgophibdidjjpildcdbfpmcahilaaf | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 27 | Avast Online Security & Privacy | gomekmidlodglbbmalcneegieacbdmki | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 28 | 喵喵折+ | ekbmhggedfdlajiikminikhcjffbleac | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 29 | 金山词霸右键查询 | nggghhclpmecenfkhdflgaaojdacpphj | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 30 | 店侦探&看店宝-淘宝卖家数据分析工具 | mndpomcndmohjljbimojnnfnennikjmk | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 31 | Web Developer | bfbameneiokkgbdmiekhjnmfkcnldhhm | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 32 | EndNote Click - Formerly Kopernio | fjgncogppolhfdpijihbpfmeohpaadpc | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 33 | 搜券助手 | kjkekpkmfhfiokgkjmcdljfomhknpphj | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 34 | Microsoft Power Automate (Legacy) | gjgfobnenmnljakmhboildkafdkicala | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 35 | Fullscreen Anything | olcfgpmjldkkjdclidhcbonieibfhhdh | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 36 | 强制网页使用自定义字体 | hckjchjpkmbihoocajjpjajkggbccgee | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 37 | 划词翻译 | hkjafmjeochlepgldkncjipkjdhpipbj | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 38 | 高效网页截图编辑插件 | mdddabjhelpilpnpgondfmehhcplpiin | 裁切/水印等复用图片能力；网页全页捕获需浏览器 adapter |
| 39 | Vue.js devtools | nhdogjmejiglipccpnnnanhbledajbpd | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 40 | 卡巴斯基保护 | ahkjpbeeocnddjkakilopmfdlnjdpcdm | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 41 | Microsoft Power Automate | ljglajjnnkapghbckkcmodicjhacbfhk | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 42 | Video Speed Controller | nffaoalbilbmmfgbnbgppjihopabppdk | 已有 video.speed/audio.volume 处理本地文件；网页实时播放控制未实现 |
| 43 | 二维码网址 | apoogihgnkfpeamhldjjoebpphjcmank | 二维码生成/识别共用 A/B；当前网页读取和浮层由浏览器适配 |
| 44 | 查看源码(zvSource) | ggkbiakmiljlbbfhjajlpjgckcjanbab | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 45 | WizClipper | jfanfpmalehkemdiiebjljddhgojhfab | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 46 | Keepa - Amazon Price Tracker | neebplgakaahbhdphmkckjjcegoiijjo | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 47 | 视频下载器 - CoCoCut | gddbgllpilhpnjpkdbopahnpealaklle | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 48 | Chrome Logger | noaneddfkdjfnfdakjjmocngnfkfehhd | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 49 | Pin Them All | ndabmaflbdfldmdlccmpccenpkgklhln | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 50 | IP Domain Country Flag | mlpapfcfoakknnhkfpencomejbcecdfp | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 51 | JSONVue | chklaanhfefbnpoihckbnefhakgolnmc | JSON、编码、计算、二维码等按主清单复用；浏览器插件界面不是现有能力 |
| 52 | 网页截图 - Screenshot Extension | akgpcdalpfphjmfifkmfbpdmgdmeeaeo | 裁切/水印等复用图片能力；网页全页捕获需浏览器 adapter |
| 53 | Free Download Manager | ahmpjcflkgiildlgicmcieglgoilbfdp | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 54 | IE Tab | hehijbfgiekmjfkfjpbkbammjbdenadd | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 55 | Zotero Connector | ekhagklcjbdpajgpjgmbionohlpdbjgc | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 56 | Video DownloadHelper | lmjnegcaeklhafolokijcfjliaokphfk | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 57 | 下载Vimeo优质视频 | phpaiffimemgakmakpcehgbophkbllkf | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 58 | XPath Helper | hgimnogjllphhhkhlmebbmlgjoejdpjl | 可后续增加本地 HTML 表格抽取/XPath 查询；动态网页获取需浏览器环境 |
| 59 | OneTab | chphlpgkkbolifaimnlloiipkdnihall | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 60 | 资源检查 | pninmpjhjchbpkopdphbkngalhfipaln | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 61 | Media Hint | akipcefbjlmpbcejgdaopmmidpnjlhnb | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 62 | MozBar | eakacpaijcpapndcfffdgphdiccmpknp | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 63 | ColorZilla | bhlhnicpbhignbdhedgjhgdocnmhomnp | 颜色换算可本地实现；网页取色/测距需 DOM/浏览器接口 |
| 64 | 小乐图客(ZIG) | gfjhimhkjmipphnaminnnnjpnlneeplk | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 65 | Bitly | iabeihobmhlgpkcgjiloemdbofjbdcic | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 66 | 新榜小助手 | eelnfhegbamoiplojcfjlncmeklaljed | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 67 | 超级下单 | dllgocogbblembnhmkbjbakkfkooficl | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 68 | Wikiwand: Wikipedia Modernized | emffkefkbkpkgpdeeooapgaicgmcbolj | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 69 | The QR Code Extension | oijdcdmnjjgnnhgljmhkjlablaejfeeb | 二维码生成/识别共用 A/B；当前网页读取和浮层由浏览器适配 |
| 70 | iGive Button | igcjdamjhkmdccbmbilbpabpofenchge | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 71 | Visualping | pemhgklkefakciniebenbfclihhmmfcd | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 72 | Neater Bookmarks | ofgjggbjanlhbgaemjbkiegeebmccifi | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 73 | Text Mode | adelhekhakakocomdfejiipdnaadiiib | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 74 | SEOquake | akdgnmcogleenhbclghghlkkdndkjdjc | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 75 | Postman Interceptor | aicmkgpgakddgnaphhhpliifpcfhicfo | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 76 | User-Agent Switcher | lkmofgnohbedopheiphabfhfjgkhfcgf | A 可做 UA 解析/生成；改写浏览器网络请求仍需扩展 |
| 77 | Advanced Font Settings | caclkomlalccbpcdllchkeecicepbmbm | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 78 | Session Buddy | edacconmaakjimmfgnblocblbcdcpbko | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 79 | SmoothScroll | nbokbjkabcmbfdlbddjidfmibcpneigj | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 80 | Fine Link Selector | adcehildlgkafghlmnbnhegagachlioe | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 81 | PDF Viewer | oemmndcbldboiebfnladdacbdfmadadm | PDF 文件渲染已有；浏览器内查看、网页打印需独立适配 |
| 82 | Steep and Cheap Countdown Timer | aaompibpddahnlhaklkapjkgiajbfkhm | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 83 | Cloudy Calculator | acgimceffoceigocablmjdpebeodphgc | JSON、编码、计算、二维码等按主清单复用；浏览器插件界面不是现有能力 |
| 84 | Scraper | mbigbapnjcgaffohmbkdlecaccepngjd | 可后续增加本地 HTML 表格抽取/XPath 查询；动态网页获取需浏览器环境 |
| 85 | TinEye Reverse Image Search | haebnnbpedcbhciplfhjjkbafijpncjl | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 86 | start.me创作的新标签页 | cfmnkhhioonhiehehedmnjibmampjiab | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 87 | FeHelper(前端助手) | pkgccpejnmalmdinmhkkfafefagiiiad | JSON、编码、计算、二维码等按主清单复用；浏览器插件界面不是现有能力 |
| 88 | Dark Reader | eimadpbcbfnmbkopoojfekhnkhdbieeh | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 89 | Octotree - GitHub code tree | bkhaagjahfmjljalopjnoealnfndnagc | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 90 | 音量增加 - 声音增强器 | ogadflejmplcdhcldlloonbiekhnlopp | 已有 video.speed/audio.volume 处理本地文件；网页实时播放控制未实现 |
| 91 | 花瓣采集插件 | hfjjjidgjckffanmaalefghfbcnifhcd | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 92 | Eagle for Chrome | lieogkinebikhdchceieedcigeafdkid | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 93 | Global Speed: 视频速度控制 | jpbjcnkcffbooppibceonlgknpkniiff | 已有 video.speed/audio.volume 处理本地文件；网页实时播放控制未实现 |
| 94 | Autofill | nlmmgnhgdeffjkdckmikfpnddkbbfkkk | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 95 | Grid Ruler | joadogiaiabhmggdifljlpkclnpfncmj | 颜色换算可本地实现；网页取色/测距需 DOM/浏览器接口 |
| 96 | Bitwarden - 免费密码管理器 | nngceckbapebfimnlniiiahkandclblb | A 只提供密码生成/评估；保险库、同步、自动填充不能以此宣称等价 |
| 97 | LastPass: Free Password Manager | hdokiejnpimakedhajhdlcegeplioahd | A 只提供密码生成/评估；保险库、同步、自动填充不能以此宣称等价 |
| 98 | Tab Manager Plus for Chrome | cnkdjjdmfiffagllbiiilooaoofcoeff | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 99 | AudioMax声音增强器 | bdbedpgdcnjmnccdappdddadbcdichio | 已有 video.speed/audio.volume 处理本地文件；网页实时播放控制未实现 |
| 100 | 腾讯 CoDesign 采集插件 | oaepbcehadahoekoceooailhlgpkndej | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 101 | 翻译 | hlppekcioiicbfafmmgikkdkljnjpiao | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 102 | 好网角收藏夹 | gbapinmejfidikhakookmkghbjpngpio | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |
| 103 | 迅雷下载支持 | ncennffkjdiamlpmcbajkmaiiiddgioo | 浏览器/外部产品适配；未核实许可与运行状态，不纳入核心同等覆盖 |

## 可复核数据

- [机器可读矩阵](qq-tool-alignment-2026-09-07.json)：166 条，包含来源、分类、状态、批次、现有 ID、差距及证据层次。
- 已保留结构化逐项核对数据与来源链接；临时 HTML 抓取和生成脚本在核对后清理，未将网站介绍文案或插件二进制复制进产品。
- 基线核对轮验证（实施前）：目录 slug 唯一、166 条全部有评估、引用的现有 Operation ID 均在真实注册表、插件 ID 唯一。仅文档变更，未运行功能代码测试；CLI 目录命令已实际执行。
