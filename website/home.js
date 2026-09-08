// Illustrative workflows, never executed in the browser. Commands match CLI contracts.
const homeTasks = [
  { label: ['Read a PDF', '读取 PDF'], intent: ['Extract readable text from this paper.', '把这篇论文提取成可读的文本。'], id: 'pdf.extract-text', provider: 'Poppler', package: 'poppler', command: 'fnsh pdf extract-text paper.pdf --unwrap -o paper.txt', input: 'paper.pdf', output: 'paper.txt', note: ['Paragraph-aware text, ready for your AI.', '合并视觉换行，交给 AI 继续阅读。'] },
  { label: ['Trim a video', '剪辑视频'], intent: ['Keep 20 seconds, starting at 10 seconds.', '从第 10 秒开始，保留 20 秒视频。'], id: 'video.trim', provider: 'FFmpeg', package: 'ffmpeg', command: 'fnsh video trim input.mp4 --start 10s --duration 20s -o clip.mp4', input: 'input.mp4', output: 'clip.mp4', note: ['An existing media engine handles the edit.', '由成熟的音视频工具完成剪辑。'] },
  { label: ['Format JSON', '整理 JSON'], intent: ['Format this JSON with two-space indentation.', '把这份 JSON 按两个空格缩进格式化。'], id: 'json.format', provider: 'Go', package: '', command: 'fnsh json format data.json --indent 2', input: 'data.json', output: 'stdout', note: ['Formatted text returned directly to stdout.', '格式化文本直接输出，无须额外依赖。'] },
];
let selectedHomeTask = 0;

function taskPreview() {
  const task = homeTasks[selectedHomeTask];
  return `<div class="task-intent"><span class="intent-avatar" aria-hidden="true">↳</span><p>${t(...task.intent)}</p></div>
    <div class="execution-path"><span>${t('Your AI','你的 AI')}</span><span aria-hidden="true">→</span><strong>FinishBit</strong><span aria-hidden="true">→</span><span>${task.provider}</span></div>
    <ol class="execution-steps">
      <li><span class="execution-number">1</span><div><span class="execution-label">${t('Find a capability','找到能力')}</span><code>${task.id}</code></div><span class="execution-check" aria-hidden="true">✓</span></li>
      <li><span class="execution-number">2</span><div><span class="execution-label">${t('Inspect the contract','确认调用方式')}</span><code>fnsh describe ${task.id}</code></div><span class="execution-check" aria-hidden="true">✓</span></li>
      <li><span class="execution-number">3</span><div><span class="execution-label">${t('Reuse an implementation','调用已有实现')}</span><code>${task.provider}</code></div><span class="execution-check" aria-hidden="true">✓</span></li>
    </ol>
    <div class="task-command"><code>${esc(task.command)}</code>${copyButton(task.command)}</div>
    <div class="task-result"><span class="result-mark" aria-hidden="true">✓</span><div><strong>${task.output}</strong><span>${t(...task.note)}</span></div><a href="#/operations/${task.id}" aria-label="${t('View contract for ','查看能力契约：')}${task.id}">↗</a></div>
    <p class="demo-disclaimer">${t('Workflow example, not a live run.','工作流示例，未在浏览器中执行。')} ${task.package ? t('Requires ','需先安装 ') + task.provider + t('.','。') : t('Built into fnsh.','fnsh 内置能力。')}</p>`;
}

function homePage() {
  return `<main id="main" class="home" tabindex="-1">
    <section class="home-hero wrap">
      <div class="hero-message">
        <p class="hero-label"><span class="mini-logo" aria-hidden="true">F.</span>${t('A practical partner for your AI','给你的 AI，一个靠谱的好助手')}</p>
        <h1>${t('Fewer tokens.<br>Work, delivered.','省下 token，<br>把事情做稳。')}</h1>
        <p class="hero-description">${t('Let AI plan. Let proven tools execute. FinishBit turns everyday tasks into reusable calls, saving the code and retries.','AI 负责思考，成熟工具负责执行。<br>用一次能力调用，省去重复写代码、调试和返工。')}</p>
        <div class="hero-actions"><a class="button primary" href="#/docs/start">${t('Get started','开始使用 FinishBit')}<span aria-hidden="true">↗</span></a><a class="button" href="#/operations">${t('Explore capabilities','探索能力')}</a></div>
      </div>
      <div class="execution-workbench">
        <div class="workbench-heading"><strong>FinishBit <span>${t('in action','如何工作')}</span></strong><span class="workbench-mode">${t('Local execution','本地执行')}</span></div>
        <div class="task-switcher" role="group" aria-label="${t('Choose a workflow example','选择工作流示例')}">${homeTasks.map((task,i)=>`<button data-home-task="${i}" aria-pressed="${selectedHomeTask===i}" aria-controls="taskPreview">${t(...task.label)}</button>`).join('')}</div>
        <div id="taskPreview" aria-live="polite" aria-atomic="true">${taskPreview()}</div>
      </div>
    </section>
    <section class="provider-band" aria-label="${t('Underlying tools','底层工具')}"><div class="wrap provider-inner"><p>${t('Built on tools<br>that do the work.','稳定执行，<br>来自成熟工具。')}</p><div class="provider-names"><a href="#/operations/guides/media">FFmpeg</a><a href="#/operations/guides/documents">Pandoc</a><a href="#/operations/guides/pdf">Poppler</a><a href="#/operations/guides/pdf">pdfcpu</a><a href="#/operations/guides/images">ImageMagick</a><a href="#/operations/guides/tables">qsv</a><a href="#/operations/guides/archives">7-Zip</a></div></div></section>
    <section class="home-section wrap token-section">
      <div class="section-introduction"><h2>${t('Spend tokens on thinking.<br>Reuse the execution.','把 token 留给思考，<br>把重复执行交给工具。')}</h2><p>${t('PDFs, data, images and media do not need a newly written script for every task. Start with a capability that already exists.','读 PDF、整理数据、处理图片和视频，不必每次都从一段新脚本开始。先找到已有能力，再把任务交给它。')}</p></div>
      <div class="comparison"><div class="comparison-before"><h3>${t('Writing it from scratch','每次从头生成脚本')}</h3><p>${t('Understand the task → generate code → resolve dependencies → debug → retry','理解任务 → 生成代码 → 处理依赖 → 调试 → 重试')}</p><div class="code-fragments" aria-hidden="true"><span>import …</span><span>def extract_text(…):</span><span>try: … except: …</span><span>${t('Revise and run again…','修改，再运行一次…')}</span></div><span class="comparison-caption">${t('More task logic enters the conversation.','更多实现细节进入对话上下文。')}</span></div>
      <div class="comparison-after"><h3>${t('Calling FinishBit','有了 FinishBit')}</h3><p>${t('Understand the task → discover → inspect → execute','理解任务 → 发现能力 → 确认契约 → 执行')}</p><div class="focused-call"><span>${t('One focused capability','只载入需要的能力')}</span><code>pdf.extract-text</code><span>${t('Defined inputs. Existing implementation.','明确输入，复用实现。')}</span></div><span class="comparison-caption">${t('Less generated code. Less context to carry.','少生成重复代码，少携带无关上下文。')}</span></div></div>
      <p class="measurement-note">${t('Token savings depend on the task, model and workflow. No fixed savings percentage is claimed.','实际 token 节省量取决于任务、模型和调用流程，不承诺固定节省比例。')}</p>
    </section>
    <section class="reliability-section"><div class="wrap reliability-layout"><div><span class="reliability-symbol" aria-hidden="true">{ ✓ }</span><h2>${t('Creative thinking. Predictable execution.','思考可以发散，执行需要确定。')}</h2><p>${t('FinishBit gives your AI a consistent way to use tested implementations and mature third-party tools.','FinishBit 把经过测试的实现与成熟第三方工具，变成 AI 可以稳定调用的能力。')}</p><a class="button" href="#/docs/execute">${t('See how execution works','了解执行与输出')}</a></div><div class="reliability-points"><article><h3>${t('Reuse proven engines','复用成熟工具')}</h3><p>${t('FFmpeg handles media. Pandoc converts documents. Defined operations connect your task to an existing implementation.','FFmpeg 处理音视频，Pandoc 转换文档。通过明确的能力调用，复用已有实现。')}</p></article><article><h3>${t('Make the contract explicit','调用之前，先看清契约')}</h3><p>${t('Inspect inputs, options and required packages before execution. Install external runtimes only when needed.','提前查看输入、参数与所需依赖，大型运行时按需安装。')}</p></article><article><h3>${t('Give your AI a result it can inspect','结果与错误，都能继续处理')}</h3><p>${t('Structured results and errors help an agent decide what to do next. Your workflow can check the output.','结构化结果与错误方便 Agent 判断下一步，让工作流能够继续检查和处理输出。')}</p></article></div></div></section>
    <section class="home-section wrap home-capabilities"><div class="capability-heading"><h2>${t('An everyday toolkit for your AI.','常见任务，已有工具接手。')}</h2><a class="button" href="#/operations">${t(`Explore ${operations.length} operations`,`探索 ${operations.length} 项能力`)}<span aria-hidden="true">↗</span></a></div><div class="home-category-grid">${categories.slice(1).map(item=>`<a href="#/operations?category=${item[0]}" class="home-category"><span class="home-category-icon">${icon(categoryIcons[item[0]])}</span><h3>${t(item[1],item[2])}</h3><p>${t(...categoryDescriptions[item[0]])}</p><span class="home-category-count">${operations.filter(o=>category(o)===item[0]).length} ${t('operations','项能力')} <span aria-hidden="true">↗</span></span></a>`).join('')}<a class="home-category extension-category" href="#/docs/extensions"><span class="home-category-icon">${icon('branch')}</span><h3>${t('Bring your own tools','也能接入你的工具')}</h3><p>${t('Turn local programs into reusable Operations.','将本地程序扩展为可复用能力。')}</p><span class="home-category-count">${t('Build an extension','开发扩展')} <span aria-hidden="true">↗</span></span></a></div></section>
    <section class="home-section wrap faq-section"><div class="faq-heading"><h2>${t('Before you get started','开始之前，你可能想知道')}</h2><p>${t('A few answers to help you decide where FinishBit fits.','关于使用方式、运行环境和数据处理。')}</p></div><div class="faq-items">${[
      [t('Does FinishBit replace my AI?','FinishBit 会替代我的 AI 吗？'),t('No. Your AI understands the task and chooses the operation. FinishBit provides the execution tools. Agents that can run local commands can use the fnsh CLI.','不会。你的 AI 负责理解任务和选择能力，FinishBit 提供执行工具。能够运行本地命令的 Agent 可以使用 fnsh CLI。')],
      [t('Does it guarantee correct AI output?','能保证 AI 的输出永远正确吗？'),t('It makes execution more predictable by reusing defined implementations. An agent can still choose the wrong operation or inputs, so important outputs should be checked.','它通过复用明确的实现，让执行更可预期。AI 仍可能选择错误的能力或输入，因此重要结果仍需检查。')],
      [t('Which platforms are supported?','哪些平台可以使用？'),t('The core CLI supports Windows, macOS and Linux. External tools have platform limits; inspect the operation contract and installation guide before use.','核心 CLI 支持 Windows、macOS 和 Linux。外部工具存在平台限制，使用前请查看能力契约和安装指南。')],
      [t('Where do my files go?','文件会被上传吗？'),t('FinishBit executes locally and does not need a resident service. How your AI handles task data depends on the agent and model you use.','FinishBit 在本地执行，不需要常驻服务。你的 AI 如何处理任务数据，取决于使用的 Agent 和模型。')],
    ].map(([question,answer])=>`<details><summary>${question}<span aria-hidden="true">+</span></summary><p>${answer}</p></details>`).join('')}</div></section>
    <section class="start-section wrap"><div class="start-panel"><div class="start-heading"><span class="logo" aria-hidden="true">F.</span><h2>${t('Give your AI a head start.','让你的 AI，从已有能力开始。')}</h2><p>${t('Install fnsh. Connect your agent. Finish the task.','安装 fnsh，接入 Agent，把下一个任务做完。')}</p></div><div class="start-install">${agentInstallCard()}<div class="start-links"><a href="#/docs/start">${t('Quick start guide','查看快速开始')} ↗</a><a href="#/docs/install">${t('Install manually','手动安装')} ↗</a></div></div></div></section>
  </main>`;
}

document.addEventListener('click', event => {
  const button = event.target.closest('[data-home-task]');
  if (!button) return;
  const index = Number(button.dataset.homeTask);
  if (!Number.isInteger(index) || !homeTasks[index] || index === selectedHomeTask) return;
  selectedHomeTask = index;
  document.querySelectorAll('[data-home-task]').forEach(item => item.setAttribute('aria-pressed', String(item === button)));
  document.querySelector('#taskPreview').innerHTML = taskPreview();
});
