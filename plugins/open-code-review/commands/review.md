---
description: 使用 Open Code Review 的 Codex 主导流程；当本机 OCR 版本过旧时，自动退回为只读审查。
---

先判断本机 `ocr` 是否支持 `ocr codex prepare`。如果支持，使用 Codex 主导流程；如果不支持（例如本机只有 v1.6.6 的 legacy review/scan），不要尝试切回 OCR 的外部 LLM 流程，也不要伪造 bundle 校验结果。

## 主流程

1. 识别用户目标。
   - 工作区：`ocr codex prepare --format json`
   - 分支 / PR：`ocr codex prepare --from <base> --to <head> --format json`
   - 提交：`ocr codex prepare --commit <sha> --format json`
   - 全量扫描：`ocr codex prepare --scan --path <paths> --format json`
2. 由 Codex 读取 bundle，补充目标上下文，完成判断。
3. 如需确认输出，执行：
   - `ocr codex validate-comments --bundle <bundle.json> --comments <comments.json>`
   - `ocr codex report --bundle <bundle.json> --comments <comments.json> --format markdown`
4. 只有用户明确要求修复时，才修改文件。

## 只读回退（read-only fallback）

如果 `ocr codex prepare` 不可用：

1. 保持同一审查范围，不缩小、不换题。
2. 用 Git range 还原变更范围，结合 CodeGraph 收集证据。
3. 只输出只读审查结论，不生成 bundle，不执行校验，也不声称运行了 OCR。
4. 结果中必须明确标注：`OCR 未运行，已使用 Git range + CodeGraph 只读回退审查`。

## 不要做的事

- 不要默认改用 `ocr review` 或 `ocr scan` 的 legacy external-LLM 模式。
- 不要依赖插件自带的新二进制；这个仓库当前没有把新 `ocr` CLI 打包进插件。
- 不要把 bundle 校验、report 结果、或未运行的 OCR 过程说成已经完成。
