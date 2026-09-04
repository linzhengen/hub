#!/usr/bin/env node
/**
 * Bootstrap ADRs in a repo:
 * - create ADR directory
 * - create adr/README.md (index) using a template
 * - create first ADR: "Adopt architecture decision records"
 */

const fs = require('node:fs');
const path = require('node:path');

function die(msg) {
  process.stderr.write(`${msg}\n`);
  process.exit(1);
}

function parseArgs(argv) {
  const out = {
    repoRoot: '.',
    dir: 'docs/decisions',
    forceIndex: false,
    indexFile: null,
    firstTitle: 'Adopt architecture decision records',
    firstStatus: 'accepted',
    deciders: '',
    technicalStory: '',
    strategy: 'date',
    json: false,
  };

  for (let i = 2; i < argv.length; i++) {
    const a = argv[i];
    const next = () => {
      if (i + 1 >= argv.length) die(`Missing value for ${a}`);
      return argv[++i];
    };

    if (a === '--repo-root') out.repoRoot = next();
    else if (a === '--dir') out.dir = next();
    else if (a === '--force-index') out.forceIndex = true;
    else if (a === '--index-file') out.indexFile = next();
    else if (a === '--first-title') out.firstTitle = next();
    else if (a === '--first-status') out.firstStatus = next();
    else if (a === '--deciders') out.deciders = next();
    else if (a === '--technical-story') out.technicalStory = next();
    else if (a === '--strategy') out.strategy = next();
    else if (a === '--json') out.json = true;
    else if (a === '--help' || a === '-h') {
      process.stdout.write(
        [
          'Usage: node bootstrap_adr.js [options]',
          '',
          'Options:',
          '  --repo-root <path>     Repo root (default: .)',
          '  --dir <path>           ADR directory (default: adr)',
          '  --index-file <path>    Override index file path (relative to repo root unless absolute)',
          '  --force-index          Overwrite index file if it exists',
          '  --first-title <text>   Title for initial ADR',
          '  --first-status <text>  Status for initial ADR (default: accepted)',
          '  --strategy date|slug|auto  Filename strategy for initial ADR (default: date)',
          '  --json                 Output machine-readable JSON (default: off)',
          '',
        ].join('\n'),
      );
      process.exit(0);
    } else {
      die(`Unknown arg: ${a}`);
    }
  }

  if (!['auto', 'date', 'slug'].includes(out.strategy))
    die(`Invalid --strategy: ${out.strategy}`);
  return out;
}

function loadReadmeTemplate() {
  const skillRoot = path.resolve(__dirname, '..');
  const templatePath = path.join(
    skillRoot,
    'assets',
    'templates',
    'adr-readme.md',
  );
  if (!fs.existsSync(templatePath))
    die(`README template not found: ${templatePath}`);
  return fs.readFileSync(templatePath, 'utf8');
}

function writeIndex(indexFile, adrDirName, { force }) {
  if (fs.existsSync(indexFile) && !force) return;
  const content = loadReadmeTemplate().replaceAll('{ADR_DIR}', adrDirName);
  fs.mkdirSync(path.dirname(indexFile), { recursive: true });
  fs.writeFileSync(indexFile, `${content.trimEnd()}\n`, 'utf8');
}

function slugify(text) {
  const t = String(text || '')
    .trim()
    .toLowerCase();
  const noQuotes = t.replace(/['"`]/g, '');
  const dashed = noQuotes.replace(/[^a-z0-9]+/g, '-').replace(/-{2,}/g, '-');
  const trimmed = dashed.replace(/^-+/, '').replace(/-+$/, '');
  return trimmed || 'decision';
}

function toPosix(p) {
  return p.split(path.sep).join('/');
}

function generateFirstAdr({ title, status, date, deciders, adrDir }) {
  const deciderLine = deciders
    ? String(deciders)
        .split(',')
        .map(s => s.trim())
        .filter(Boolean)
        .join(', ')
    : '';

  return `---
status: ${status}
date: ${date}
decision-makers: ${deciderLine}
---

# ${title}

## 文脈と課題

hub の設計判断は、これまで 2 か所に散っていた。コードコメントと、
\`docs/design/\` の方向性メモである。

コードコメントは強い。実際、このリポジトリの規律の大半はそこにある
（「期限は SQL でフィルタせず判定時に落とす」「認可経路は 1 本」など）。
ただしコメントは**その場所に触る人にしか届かない**。判断を知らずに別の場所から
同じ不変条件を壊すことは防げない。

方向性メモは領域全体の物語を持つが、1 つの判断だけを指して
「これは受理済みか」「いつ覆すのか」「誰が決めたのか」を答える形にはなっていない。

足りないのは、**1 判断を 1 ファイルで、状態と帰結ごと**持つ場所である。

## 判断

\`${adrDir}/\` に ADR を置く。MADR 4.0 に準拠しつつ、実装計画と検証手順を足す。

- 1 ファイル 1 判断。ファイル名は \`YYYY-MM-DD-english-slug.md\`、**本文は日本語**
  （\`docs/design/\` と同じ流儀）
- \`proposed\` で始まり、\`accepted\` か \`rejected\` に移る
- 置き換えるときは新しい ADR を作り、双方向にリンクする
- **自己完結させる。**エージェントが 1 本読んで、追加の質問なしに実装を始められること

### 非目標

- \`docs/design/\` の方向性メモを置き換えない。役割が違う（物語 対 1 判断）
- 過去の判断を遡って ADR 化しない。争点になったものから昇格させる
- 公開しない。\`docs/build.mjs\` は \`docs/site/\` しか読まない

## 帰結

- 良い点: 判断がコードと同じリポジトリに、版管理されて残る
- 良い点: 後から来る人やエージェントが「なぜ」を辿れる。特に、既に守られている
  不変条件を知らずに壊すことを止められる
- 良い点: 一度決めたことを蒸し返さずに済む
- 悪い点: 書くのに時間がかかる。ただし良い ADR は、かけた時間より多くを節約する
- どちらでもない点: 古くなった ADR を \`deprecated\` にする手入れが要る

## 却下した案

- **記録を作らない**: 会話とコードコメントのまま。コメントはその場所に触る人にしか
  届かず、判断の状態（受理済みか、覆されたか）も持てない。
- **リポジトリ外の Wiki**: コードと乖離し、版管理もされない。このリポジトリが
  生成物と規約をすべてコードの中に置いているのと矛盾する。
- **方向性メモに書き続ける**: メモは複数の判断をまたぐ物語で、1 判断の状態を
  持つようにはできていない。両方に本体を置くと 2 か所で食い違う。

## 補足

- MADR: <https://adr.github.io/madr/>
- Michael Nygard, "Documenting Architecture Decisions":
  <https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions>
- 規約: \`.agents/skills/adr-skill/references/adr-conventions.md\``;
}

function updateIndexFile(indexFile, { relLink, title, status, date }) {
  if (!fs.existsSync(indexFile)) return;
  let content = fs.readFileSync(indexFile, 'utf8');
  if (content.includes(relLink)) return;

  const entryLine = `- [${title}](${relLink}) (${status}, ${date})`;

  // Append after "## ADRs" heading if found, otherwise append at end
  const normalized = content.replace(/\r\n/g, '\n');
  const lines = normalized.split('\n');
  const headingIdx = lines.findIndex(l => /^##\s+ADRs\s*$/i.test(l));

  if (headingIdx !== -1) {
    // Insert after the heading (and any blank line after it)
    let insertAt = headingIdx + 1;
    while (insertAt < lines.length && lines[insertAt].trim() === '') insertAt++;
    lines.splice(insertAt, 0, entryLine);
  } else {
    lines.push(entryLine);
  }

  fs.writeFileSync(indexFile, lines.join('\n'), 'utf8');
}

function main() {
  const args = parseArgs(process.argv);

  const repoRoot = path.resolve(process.cwd(), args.repoRoot);
  if (!fs.existsSync(repoRoot)) die(`Repo root does not exist: ${repoRoot}`);

  const adrDir = path.resolve(repoRoot, args.dir);
  fs.mkdirSync(adrDir, { recursive: true });

  const indexFile = args.indexFile
    ? path.isAbsolute(args.indexFile)
      ? args.indexFile
      : path.resolve(repoRoot, args.indexFile)
    : path.join(adrDir, 'README.md');

  const indexExistedBefore = fs.existsSync(indexFile);
  writeIndex(indexFile, args.dir, { force: args.forceIndex });
  const indexWritten =
    fs.existsSync(indexFile) && (!indexExistedBefore || args.forceIndex);

  // Create the first ADR as a filled-out decision (not a blank template).
  const relIndex = path.isAbsolute(indexFile)
    ? path.relative(repoRoot, indexFile)
    : indexFile;
  const today = new Date().toISOString().slice(0, 10);

  const firstAdrContent = generateFirstAdr({
    title: args.firstTitle,
    status: args.firstStatus,
    date: today,
    deciders: args.deciders,
    adrDir: args.dir,
  });

  // Determine filename using same logic as new_adr.js
  const strategy = args.strategy === 'auto' ? 'date' : args.strategy;
  let firstAdrFilename;
  if (strategy === 'date') {
    firstAdrFilename = `${today}-${slugify(args.firstTitle)}.md`;
  } else {
    firstAdrFilename = `${slugify(args.firstTitle)}.md`;
  }
  const firstAdrPath = path.join(adrDir, firstAdrFilename);
  fs.writeFileSync(firstAdrPath, `${firstAdrContent.trimEnd()}\n`, 'utf8');

  // Update index
  const relLink = toPosix(path.relative(path.dirname(indexFile), firstAdrPath));
  updateIndexFile(indexFile, {
    relLink,
    title: args.firstTitle,
    status: args.firstStatus,
    date: today,
  });

  if (args.json) {
    const payload = {
      repoRoot,
      adrDir,
      adrDirRelPath: toPosix(path.relative(repoRoot, adrDir)),
      indexPath: indexFile,
      indexRelPath: toPosix(relIndex),
      indexExistedBefore,
      indexWritten,
      firstAdr: {
        createdAdrPath: firstAdrPath,
        createdAdrRelPath: toPosix(path.relative(repoRoot, firstAdrPath)),
        title: args.firstTitle,
        status: args.firstStatus,
        strategy,
        date: today,
      },
      date: today,
    };
    process.stdout.write(`${JSON.stringify(payload)}\n`);
    return;
  }

  process.stdout.write(`${firstAdrPath}\n`);
  process.stdout.write(`Bootstrapped ADRs at ${adrDir} (${today})\n`);
  process.stdout.write(`Index: ${indexFile}\n`);
}

main();
