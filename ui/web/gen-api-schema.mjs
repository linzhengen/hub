import { execFile } from 'node:child_process';
import { mkdir, readdir } from 'node:fs/promises';
import { extname, join, relative } from 'node:path';
import process from 'node:process';
import { promisify } from 'node:util'; // 生成される TypeScript の出力ディレクトリ

const execFileAsync = promisify(execFile);

const inputDir = '../../server/openapi'; // OpenAPI YAML のディレクトリ
const outputDir = './src/api/schema';

async function findYamlFiles(dir) {
  const results = [];
  const entries = await readdir(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = join(dir, entry.name);
    if (entry.isDirectory()) {
      results.push(...(await findYamlFiles(fullPath)));
    } else if (extname(entry.name) === '.yaml') {
      results.push(fullPath);
    }
  }

  return results;
}

async function generateSchema() {
  const yamlFiles = await findYamlFiles(inputDir);
  if (yamlFiles.length === 0) {
    console.error('No OpenAPI YAML files found.');
    process.exit(1);
  }

  await mkdir(outputDir, { recursive: true });

  for (const yamlFile of yamlFiles) {
    // `inputDir` からの相対パスを取得し、スラッシュを `_` に置換してファイル名を作成
    const relativePath = relative(inputDir, yamlFile);
    const fileName = relativePath
      .replaceAll('/', '-')
      .replace(/\.yaml$/, '.ts');
    const outputPath = join(outputDir, fileName);


    console.info(
      `Generating TypeScript models for: ${yamlFile} → ${outputPath}`,
    );

    try {
      // The generator is a devDependency rather than an `npx` fetch: its
      // output depends on its version, so an unpinned one makes a proto change
      // rewrite files it did not touch. Same reasoning as the pinned protoc
      // plugins in the root Makefile.
      await execFileAsync(
        'pnpm',
        ['exec', 'openapi-typescript', yamlFile, '--output', outputPath],
        { stdio: 'inherit' },
      );
    } catch (error) {
      console.error(`Failed to generate models for ${yamlFile}:`, error);
    }
  }

  console.log('✅ TypeScript models generation completed!');
}

await generateSchema();
