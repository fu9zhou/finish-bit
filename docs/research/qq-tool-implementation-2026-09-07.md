# QQ工具逐项实施状态（2026-09-08）

基于[原始166项核对](qq-tool-alignment-2026-09-07.md)，记录本轮真实实现边界。注册表为277个操作，新增107个；不是166项全部同等覆盖。基础已实现表示所述最小本地能力已落地，子集与平台限制单列。

| 状态 | 网站入口数 |
| --- | ---: |
| 子集已实现 | 24 |
| 专项待评估 | 19 |
| 保留已有 | 23 |
| 基础已实现 | 52 |
| 待用户决定 | 6 |
| 词库暂缓 | 19 |
| 平台子集 | 2 |
| 仅通用OCR基础 | 17 |
| 服务边界 | 4 |

| # | 网站入口 | 实施状态 | 可调用操作 | 边界 |
| ---: | --- | --- | --- | --- |
| 1 | [图片压缩](https://tool.browser.qq.com/tupianyasuo.html) | 子集已实现 | `image.target-size`, `image.compress` | JPEG 目标字节数；无法达到时失败。批量可逐文件调用。 |
| 2 | [证件照生成](https://tool.browser.qq.com/id_photo.html) | 子集已实现 | `image.id-photo` | 标准像素尺寸、中心裁切、已有透明背景填色；无自动抠图/人脸对齐。 |
| 3 | [PDF转Word](https://tool.browser.qq.com/pdf_2_word.html) | 子集已实现 | `pdf.to-docx`, `document.scan` | 提取文字或图片 OCR 后重排 DOCX；无复杂版式重建。 |
| 4 | [PDF转Excel](https://tool.browser.qq.com/pdf_2_excel.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 5 | [PDF转HTML](https://tool.browser.qq.com/pdf_2_html.html) | 保留已有 | `pdf.to-html` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 6 | [PDF转图片](https://tool.browser.qq.com/pdf_2_png.html) | 基础已实现 | `pdf.render`, `pdf.to-long-image` | 分页图与长图；显式选页，最多50页。 |
| 7 | [PDF转PPT](https://tool.browser.qq.com/pdf_2_ppt.html) | 子集已实现 | `pdf.to-pptx` | 一页一张图片幻灯片；文字不可编辑。 |
| 8 | [Word转PDF](https://tool.browser.qq.com/word_2_pdf.html) | 待用户决定 |  | LibreOffice MSI展开约1.47 GiB，超出1GB阈值；未新增完整Office渲染。 |
| 9 | [垃圾分类查询](https://tool.browser.qq.com/garbage.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 10 | [Word格式转换](https://tool.browser.qq.com/word_convert.html) | 待用户决定 |  | LibreOffice MSI展开约1.47 GiB，超出1GB阈值；未新增完整Office渲染。 |
| 11 | [PPT格式转换](https://tool.browser.qq.com/ppt_convert.html) | 待用户决定 |  | LibreOffice MSI展开约1.47 GiB，超出1GB阈值；未新增完整Office渲染。 |
| 12 | [PDF加水印](https://tool.browser.qq.com/pdf_watermark.html) | 保留已有 | `pdf.watermark`, `pdf.stamp` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 13 | [PDF瘦身](https://tool.browser.qq.com/pdf_compress.html) | 子集已实现 | `pdf.compress-images`, `pdf.optimize` | JPEG 重新栅格化和结构优化；返回大小，无任意目标字节保证。 |
| 14 | [在线录屏](https://tool.browser.qq.com/screen_record.html) | 平台子集 | `screen.record` | Windows FFmpeg 桌面/矩形无声 MP4；仅参数验证，未实际录屏。 |
| 15 | [Excel转PDF](https://tool.browser.qq.com/excel_to_pdf.html) | 待用户决定 |  | LibreOffice MSI展开约1.47 GiB，超出1GB阈值；未新增完整Office渲染。 |
| 16 | [去手写](https://tool.browser.qq.com/handwriting_erasure.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 17 | [文档瘦身](https://tool.browser.qq.com/office_reduce.html) | 基础已实现 | `document.compress` | OOXML ZIP和图片减重；签名包拒绝、元数据图片保留、不变小则原样副本。 |
| 18 | [身份证识别](https://tool.browser.qq.com/identification.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 19 | [银行卡识别](https://tool.browser.qq.com/bankcard_ocr.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 20 | [印刷体识别](https://tool.browser.qq.com/basic_ocr.html) | 基础已实现 | `ocr.text`, `ocr.words`, `ocr.to-pdf`, `ocr.to-html` | 中英文印刷体；已真实引擎验收，不承诺通用识别准确率。 |
| 21 | [手写体识别](https://tool.browser.qq.com/handwriting_ocr.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 22 | [广告体识别](https://tool.browser.qq.com/ocr_advertise.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 23 | [英语识别](https://tool.browser.qq.com/ocr_english.html) | 基础已实现 | `ocr.text`, `ocr.words` | 英文印刷体已验收。 |
| 24 | [护照识别](https://tool.browser.qq.com/passport_ocr.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 25 | [PDF签名](https://tool.browser.qq.com/pdf_sign.html) | 子集已实现 | `pdf.sign-image` | 用户提供签名图片，位置/尺寸/页选择；非证书数字签名。 |
| 26 | [运单识别](https://tool.browser.qq.com/waybill.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 27 | [集装箱识别](https://tool.browser.qq.com/container_ocr.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 28 | [车辆证件识别](https://tool.browser.qq.com/carcard_ocr.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 29 | [营业执照识别](https://tool.browser.qq.com/ocr_bizLicense.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 30 | [名片识别](https://tool.browser.qq.com/ocr_businesscard.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 31 | [港澳台证件识别](https://tool.browser.qq.com/ocr_permit.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 32 | [银行回单识别](https://tool.browser.qq.com/bankslip_ocr.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 33 | [完税证明识别](https://tool.browser.qq.com/ocr_dutypaidproof.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 34 | [医疗票据识别](https://tool.browser.qq.com/ocr_recognize_medical_invoice.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 35 | [网约车行程单识别](https://tool.browser.qq.com/onlinetaxi_ocr.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 36 | [人脸年龄选择](https://tool.browser.qq.com/face_age_transformation.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 37 | [图片编辑器](https://tool.browser.qq.com/img_edit_canvas.html) | 子集已实现 | `image.pixelate`, `image.filter`, `image.crop`, `image.annotate` | 本地编辑操作；未开发交互画布。 |
| 38 | [密码安全检测](https://tool.browser.qq.com/password_check.html) | 基础已实现 | `password.inspect` | 本地长度、字符和重复模式检查；不联网查询泄漏库。 |
| 39 | [文件安全检测](https://tool.browser.qq.com/file_scan.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 40 | [手机安装包检测](https://tool.browser.qq.com/app_inspector.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 41 | [搜狗百宝箱](https://tool.browser.qq.com/sogou_box.html) | 服务边界 |  | 在线服务或浏览器/账户集成超出本地算法；未接第三方处理API。 |
| 42 | [合同验签](https://tool.browser.qq.com/contract_verification.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 43 | [合同对比](https://tool.browser.qq.com/contract_comparison.html) | 子集已实现 | `document.compare` | 提取文字逐行diff；无视觉版式对比或法律判断。 |
| 44 | [图片转PDF](https://tool.browser.qq.com/img_2_pdf_convert.html) | 保留已有 | `pdf.from-images` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 45 | [PPT转PDF](https://tool.browser.qq.com/ppt_2_pdf.html) | 待用户决定 |  | LibreOffice MSI展开约1.47 GiB，超出1GB阈值；未新增完整Office渲染。 |
| 46 | [搜狗PDF文件搜索](https://tool.browser.qq.com/pdf_search.html) | 服务边界 |  | 在线服务或浏览器/账户集成超出本地算法；未接第三方处理API。 |
| 47 | [PDF拆分](https://tool.browser.qq.com/pdf_split.html) | 保留已有 | `pdf.split`, `pdf.select-pages` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 48 | [PDF合并](https://tool.browser.qq.com/pdf_merge.html) | 保留已有 | `pdf.merge` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 49 | [转纯图PDF](https://tool.browser.qq.com/pdf_imagefy.html) | 基础已实现 | `pdf.rasterize` | 统一选页/DPI的纯图 PDF；不保留原文字和交互对象。 |
| 50 | [PDF页面管理](https://tool.browser.qq.com/pdf_page_manage.html) | 保留已有 | `pdf.select-pages`, `pdf.remove-pages`, `pdf.insert-pages`, `pdf.rotate` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 51 | [PDF图片提取](https://tool.browser.qq.com/pdf_img_extract.html) | 保留已有 | `pdf.extract-images` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 52 | [PDF加解密](https://tool.browser.qq.com/pdf_password.html) | 保留已有 | `pdf.encrypt`, `pdf.decrypt` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 53 | [PDF页面裁剪](https://tool.browser.qq.com/pdf_crop.html) | 基础已实现 | `pdf.crop` | 四边独立非负点数边距，支持页选择。 |
| 54 | [修改PDF元数据](https://tool.browser.qq.com/pdf_metadata.html) | 基础已实现 | `pdf.metadata-set` | 标题/作者/主题/创建者；不修改时间戳等所有底层字段。 |
| 55 | [字帖生成](https://tool.browser.qq.com/zitie_new.html) | 基础已实现 | `worksheet.handwriting` | 可离线打印的HTML方格/田字格/米字格描红，使用系统字体。 |
| 56 | [Excel格式转换](https://tool.browser.qq.com/excel_convert.html) | 待用户决定 |  | LibreOffice MSI展开约1.47 GiB，超出1GB阈值；未新增完整Office渲染。 |
| 57 | [图片转Excel](https://tool.browser.qq.com/table_recognize.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 58 | [今天吃什么](https://tool.browser.qq.com/whattoeat.html) | 基础已实现 | `random.choose` | 用户菜单随机抽取/洗牌；无内置菜单或种子模式。 |
| 59 | [文本转语音](https://tool.browser.qq.com/tts.html) | 平台子集 | `speech.voices`, `speech.synthesize` | Windows已安装离线声音→WAV，已验收；不另下载声音。 |
| 60 | [提取文字](https://tool.browser.qq.com/ocr.html) | 基础已实现 | `ocr.text`, `pdf.ocr-text` | 本地图片与扫描PDF文字提取。 |
| 61 | [亲戚关系计算](https://tool.browser.qq.com/relatives_name.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 62 | [PDF加页码](https://tool.browser.qq.com/pdf_page_number.html) | 基础已实现 | `pdf.page-numbers` | 动态当前页、总页、偏移模板和位置。 |
| 63 | [修改PDF页面尺寸](https://tool.browser.qq.com/pdf_pagesize.html) | 基础已实现 | `pdf.resize-pages` | A3/A4/A5/Letter/Legal及方向，真实引擎验收。 |
| 64 | [高校查询](https://tool.browser.qq.com/school.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 65 | [图片转字符画](https://tool.browser.qq.com/img_2_text.html) | 基础已实现 | `image.ascii` | 灰度字符画、宽度和自定义字符渐变。 |
| 66 | [人脸性别转换](https://tool.browser.qq.com/face_gender_transformation.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 67 | [卡路里查询](https://tool.browser.qq.com/calories.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 68 | [增加图片大小](https://tool.browser.qq.com/img_enlarge.html) | 基础已实现 | `image.minimum-bytes` | PNG合法辅助块增加字节数；不增加真实细节。 |
| 69 | [发票提取](https://tool.browser.qq.com/invoice_extract.html) | 仅通用OCR基础 | `ocr.text`, `ocr.words` | 已可提取普通印刷文字及坐标；此类证件/手写/票据/表格专用解析尚未实现或验收。 |
| 70 | [颜色转换](https://tool.browser.qq.com/colortrans.html) | 子集已实现 | `color.convert` | HEX/RGB输入，输出HEX/RGB/HSL/HSV/CMYK；无ICC。 |
| 71 | [图片日漫滤镜](https://tool.browser.qq.com/img_anime_filter.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 72 | [头像二次元](https://tool.browser.qq.com/img_face_anime.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 73 | [美化你的头像](https://tool.browser.qq.com/img_module_face_stylize.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 74 | [油画滤镜](https://tool.browser.qq.com/img_painting_filter.html) | 基础已实现 | `image.filter` | 确定性油画/素描/棕褐/炭笔/负片/自动调色预设。 |
| 75 | [人像分割](https://tool.browser.qq.com/img_human_split.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 76 | [照片美化](https://tool.browser.qq.com/img_pic_beauty.html) | 子集已实现 | `image.filter`, `image.adjust` | 普通调色滤镜；无语义人像美颜。 |
| 77 | [文档图片扭曲恢复](https://tool.browser.qq.com/doc_wrap.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 78 | [图片修复](https://tool.browser.qq.com/img_fix.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 79 | [图片去摩尔纹](https://tool.browser.qq.com/image_enhance.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 80 | [多人卡通画](https://tool.browser.qq.com/img_ai_cartoon.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 81 | [童话脸](https://tool.browser.qq.com/img_ai_face.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 82 | [乔丹风格化](https://tool.browser.qq.com/img_ai_jordanstyle.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 83 | [国漫水彩风](https://tool.browser.qq.com/img_ai_watercolor.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 84 | [头像挂饰](https://tool.browser.qq.com/avatar_pendant.html) | 子集已实现 | `image.watermark`, `image.canvas`, `image.crop` | 复用用户提供挂饰素材；无站点模板。 |
| 85 | [字数计算](https://tool.browser.qq.com/wordcount.html) | 保留已有 | `text.count` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 86 | [英文创业公司/项目名生成](https://tool.browser.qq.com/startupname.html) | 基础已实现 | `name.generate` | 用户前后缀词表组合，不保证商标可用或不重复。 |
| 87 | [手持弹幕LED](https://tool.browser.qq.com/led.html) | 基础已实现 | `text.led` | 离线HTML滚动文字，支持减少动态效果偏好。 |
| 88 | [JSON diff](https://tool.browser.qq.com/jsondiff.html) | 基础已实现 | `json.diff` | JSON Pointer结构差异，缺失和null区分、精确有界数字。 |
| 89 | [JSON校验](https://tool.browser.qq.com/jsoncheck.html) | 保留已有 | `json.validate` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 90 | [文本颜艺](https://tool.browser.qq.com/emoji.html) | 基础已实现 | `text.emoticons` | 自建10个简单文本表情，不复制外部词库。 |
| 91 | [古诗词取名](https://tool.browser.qq.com/makename.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 92 | [车牌归属地](https://tool.browser.qq.com/carnumber.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 93 | [电话区号查询](https://tool.browser.qq.com/phonenumber.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 94 | [英文人名生成](https://tool.browser.qq.com/fakeword.html) | 子集已实现 | `name.generate` | 用户提供英文名字词表；无预装人名库。 |
| 95 | [温度转换](https://tool.browser.qq.com/temperaturetrans.html) | 基础已实现 | `temperature.convert` | C/F/K换算与绝对零度校验。 |
| 96 | [投资收益计算](https://tool.browser.qq.com/invest.html) | 基础已实现 | `finance.compound` | 用户输入利率、复利频率、定期追加；无收益预测。 |
| 97 | [血型遗传规律](https://tool.browser.qq.com/bloodtype.html) | 基础已实现 | `biology.blood-types` | 简化ABO遗传可能性枚举；不用于亲子鉴定。 |
| 98 | [五险一金计算](https://tool.browser.qq.com/wuxianyijin.html) | 子集已实现 | `finance.contributions` | 用户提供基数与比例；无城市实时政策库。 |
| 99 | [md5加密](https://tool.browser.qq.com/md5.html) | 保留已有 | `hash.calculate` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 100 | [进制转换](https://tool.browser.qq.com/hexconvert.html) | 基础已实现 | `number.base` | 2至36进制任意精度整数。 |
| 101 | [简体繁体转换](https://tool.browser.qq.com/chinese.html) | 基础已实现 | `text.chinese` | OpenCC 1.1.7字词典，最长词优先；非完整mmseg或地区变体。 |
| 102 | [房贷计算](https://tool.browser.qq.com/mortgage.html) | 基础已实现 | `finance.mortgage` | 等额本息/等额本金和计划表；参数由用户给定。 |
| 103 | [Unicode编解码](https://tool.browser.qq.com/unicode.html) | 基础已实现 | `unicode.encode`, `unicode.decode` | JSON Unicode转义和代理对；拒绝孤立代理项。 |
| 104 | [url编解码](https://tool.browser.qq.com/urlencode.html) | 保留已有 | `url.encode`, `url.decode` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 105 | [数字大小写转换](https://tool.browser.qq.com/num2zh.html) | 基础已实现 | `number.chinese` | 中文数字及人民币大写，定点角分；不隐式舍入更多小数。 |
| 106 | [BMI计算](https://tool.browser.qq.com/bmi.html) | 基础已实现 | `health.bmi` | 仅BMI数值，无未经维护的分类建议。 |
| 107 | [IP归属地](https://tool.browser.qq.com/iplocation.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 108 | [文本比较](https://tool.browser.qq.com/textdiff.html) | 基础已实现 | `text.diff` | 精确逐行LCS，每边最多2000行。 |
| 109 | [JSON格式化](https://tool.browser.qq.com/jsonbeautify.html) | 保留已有 | `json.format`, `json.minify` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 110 | [时间戳转换](https://tool.browser.qq.com/timestamp.html) | 保留已有 | `time.convert` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 111 | [在线加解密](https://tool.browser.qq.com/crypto.html) | 子集已实现 | `crypto.key`, `crypto.encrypt`, `crypto.decrypt` | AES-256-GCM文本信封，需256位密钥；非所有网站兼容算法。 |
| 112 | [GIF分解](https://tool.browser.qq.com/gifsplitter.html) | 保留已有 | `image.gif-split` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 113 | [URL解析](https://tool.browser.qq.com/urlparse.html) | 基础已实现 | `url.parse` | 保留重复query参数及URL各组成部分。 |
| 114 | [历史朝代查询](https://tool.browser.qq.com/dynasties.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 115 | [各国首都](https://tool.browser.qq.com/capital.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 116 | [日期计算](https://tool.browser.qq.com/datecal.html) | 基础已实现 | `date.add`, `date.diff`, `date.business-days`, `date.age` | 月末截断；工作日使用用户调休表。 |
| 117 | [计算器](https://tool.browser.qq.com/calculator.html) | 基础已实现 | `math.evaluate`, `math.statistics` | 白名单数学表达式和有限数统计；不执行脚本。 |
| 118 | [随机密码生成](https://tool.browser.qq.com/pwdgenerator.html) | 基础已实现 | `password.generate` | crypto/rand独立均匀字符；非强制类别配额。 |
| 119 | [长度转换](https://tool.browser.qq.com/lengthconvert.html) | 基础已实现 | `unit.convert` | 长度/面积/体积/质量/时长/数据的精确比率。 |
| 120 | [随机数生成](https://tool.browser.qq.com/random.html) | 基础已实现 | `random.integer`, `random.choose` | 密码学随机区间与无放回抽样；无种子复现。 |
| 121 | [base64编码](https://tool.browser.qq.com/base64.html) | 保留已有 | `base64.encode`, `base64.decode` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 122 | [二维码生成](https://tool.browser.qq.com/qrcode.html) | 基础已实现 | `qrcode.generate` | 纠错/尺寸/配色，输出前回扫。 |
| 123 | [视频转gif](https://tool.browser.qq.com/video_2_gif.html) | 保留已有 | `video.gif` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 124 | [markdown编辑器](https://tool.browser.qq.com/markdown.html) | 子集已实现 | `markdown.to-html` | GFM本地HTML，禁原始HTML和远程资源；非交互编辑器。 |
| 125 | [九宫格切图](https://tool.browser.qq.com/img9grid.html) | 基础已实现 | `image.split-grid` | NxM行优先切图，非整除边缘完整覆盖。 |
| 126 | [图片格式转换](https://tool.browser.qq.com/imgconvert.html) | 保留已有 | `image.convert`, `image.formats` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 127 | [GIF合成](https://tool.browser.qq.com/gifcreate.html) | 保留已有 | `image.gif-create` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 128 | [文字转拼音](https://tool.browser.qq.com/tta.html) | 基础已实现 | `text.pinyin` | 拼音/声调/首字母/多音候选；无上下文消歧。 |
| 129 | [正则校验](https://tool.browser.qq.com/regexp.html) | 基础已实现 | `regex.validate`, `regex.match`, `regex.replace` | 新建Go RE2接口；原text.replace仅字面替换。 |
| 130 | [二维码美化](https://tool.browser.qq.com/prettify_qrcode.html) | 子集已实现 | `qrcode.generate` | 配色和尺寸美化并回扫；无图标嵌入或艺术模型。 |
| 131 | [二维码名片](https://tool.browser.qq.com/visit_card.html) | 基础已实现 | `qrcode.contact` | vCard QR，字段转义与输出回扫。 |
| 132 | [快递查询](https://tool.browser.qq.com/kuaidi.html) | 服务边界 |  | 在线服务或浏览器/账户集成超出本地算法；未接第三方处理API。 |
| 133 | [成语接龙](https://tool.browser.qq.com/jielong.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 134 | [医院推荐](https://tool.browser.qq.com/hospitalrecommend.html) | 服务边界 |  | 在线服务或浏览器/账户集成超出本地算法；未接第三方处理API。 |
| 135 | [便捷思维导图](https://tool.browser.qq.com/markmap.html) | 子集已实现 | `markdown.mindmap` | 离线可折叠大纲HTML；非自由布局编辑器。 |
| 136 | [汉字标准发音](https://tool.browser.qq.com/hanzifayin.html) | 子集已实现 | `text.pinyin`, `speech.synthesize` | 拼音标注/用户已有系统声音；不等于权威字典逐字录音。 |
| 137 | [YAML/JSON互相转换](https://tool.browser.qq.com/yaml_2_json.html) | 基础已实现 | `yaml.to-json`, `json.to-yaml`, `yaml.format`, `yaml.validate` | 单文档、字符串键、拒绝别名和重复键、保留有界数值。 |
| 138 | [图片黑白化](https://tool.browser.qq.com/img_fade.html) | 保留已有 | `image.grayscale` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 139 | [元素周期表](https://tool.browser.qq.com/periodic.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 140 | [食物热量表](https://tool.browser.qq.com/calories_list.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 141 | [卡路里查询](https://tool.browser.qq.com/food_calories.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 142 | [火星文翻译器](https://tool.browser.qq.com/toMars.html) | 子集已实现 | `text.map` | 用户字符映射一次转换，非标准火星文词库。 |
| 143 | [翻译](https://tool.browser.qq.com/translate.html) | 专项待评估 |  | 需要专项引擎/模型/样本；现有普通变换不等于该能力。详见原评估方案。 |
| 144 | [汉字偏旁](https://tool.browser.qq.com/radical.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 145 | [歇后语](https://tool.browser.qq.com/allegory.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 146 | [词语注解](https://tool.browser.qq.com/explain.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 147 | [世界时间校准](https://tool.browser.qq.com/timer.html) | 基础已实现 | `time.world` | 内置IANA时区显示同一时刻，不校准系统时钟。 |
| 148 | [成语大全](https://tool.browser.qq.com/chengyujielong.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 149 | [随机网名](https://tool.browser.qq.com/nick.html) | 基础已实现 | `name.generate` | 用户词表组合随机网名。 |
| 150 | [快递信息提取](https://tool.browser.qq.com/number_acquisition.html) | 子集已实现 | `contact.extract` | 电话/邮箱/单号候选；无地址姓名行政区自动解析。 |
| 151 | [字节数换算](https://tool.browser.qq.com/byte_cal.html) | 基础已实现 | `unit.convert` | SI与IEC、bit/byte明确区分。 |
| 152 | [邮编查询](https://tool.browser.qq.com/zipcode.html) | 词库暂缓 |  | 用户将此类查询设为低优先级；缺少已验收可再分发数据，不复制网站数据。 |
| 153 | [userAgent工具](https://tool.browser.qq.com/useragent.html) | 子集已实现 | `useragent.parse` | 常见浏览器/系统token启发式；保留原文。 |
| 154 | [保质期计算](https://tool.browser.qq.com/shelflife.html) | 基础已实现 | `date.expiry` | 自然月/天期限、明确参考日期与到期边界。 |
| 155 | [取名字](https://tool.browser.qq.com/naming.html) | 子集已实现 | `name.generate` | 用户名字素材；无内置专业取名库。 |
| 156 | [文本去重](https://tool.browser.qq.com/unique.html) | 保留已有 | `text.unique` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 157 | [图片像素化](https://tool.browser.qq.com/img_pixel.html) | 基础已实现 | `image.pixelate` | 整图或矩形区域块平均马赛克。 |
| 158 | [硬盘分区](https://tool.browser.qq.com/partition.html) | 基础已实现 | `disk.capacity` | 容量显示/对齐计算；不修改磁盘。 |
| 159 | [图片隐写](https://tool.browser.qq.com/image_secret_msg.html) | 基础已实现 | `image.hide`, `image.reveal` | PNG RGB LSB、长度/版本/CRC32校验。 |
| 160 | [开发语言输出hello world](https://tool.browser.qq.com/compilation.html) | 基础已实现 | `code.hello` | 12种语言源码模板；不安装或执行解释器。 |
| 161 | [文字隐写](https://tool.browser.qq.com/text_secret_msg.html) | 基础已实现 | `text.hide`, `text.reveal` | 零宽编码、版本、CRC32；文字清洗可能破坏载荷。 |
| 162 | [uuid生成](https://tool.browser.qq.com/uuid.html) | 保留已有 | `uuid.generate` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 163 | [图片加水印](https://tool.browser.qq.com/watermark.html) | 保留已有 | `image.watermark` | 复用原有操作；本轮没有宣称与网站逐样本全等价。 |
| 164 | [表情包制作](https://tool.browser.qq.com/biaoqing.html) | 子集已实现 | `image.annotate`, `image.canvas`, `image.gif-create` | 复用用户图片/字体的文字和GIF操作；无现成表情模板。 |
| 165 | [二维码扫描](https://tool.browser.qq.com/qrcode_scan.html) | 基础已实现 | `qrcode.decode` | 本地PNG/JPEG/GIF二维码识别，无摄像头捕获。 |
| 166 | [文档扫描](https://tool.browser.qq.com/word_scan.html) | 子集已实现 | `document.scan`, `pdf.ocr` | 图像或PDF OCR重排文档；无自动透视/曲面展平。 |

安装、使用、依赖体积和验证命令见[本地工具说明](../local-tools.md)。JSON矩阵保留最初status/current/plan字段，实施后状态位于implementation字段；读者应使用后者判断本轮覆盖。
