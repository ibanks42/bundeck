import { sync } from "glob";
import path from "node:path";

const projectRoot = path.join(__dirname, "..");
const join = (...paths) => path.join(projectRoot, ...paths);

const patterns = [join("src/**/*.{js,ts,jsx,tsx,json}")];
const batchSize = 50; // Adjust the batch size if needed

const options = {
	cwd: projectRoot,
	ignore: [
		join("node_modules/**"),
		join("dist/**"),
		join("src/components/ui/**"),
		join("src/routeTree.gen.ts"),
	],
};

const files = patterns.flatMap((pattern) => sync(pattern, options));

const runBatch = async (fileBatch, callback) => {
	const exec = Bun.spawn([
		join("node_modules/@biomejs/biome/bin/biome"),
		"check",
		"--formatter-enabled=true",
		"--linter-enabled=true",
		"--organize-imports-enabled=true",
		"--write",
		...fileBatch,
	]);

	const stdout = await new Response(exec.stdout).text();
	const error = await new Response(exec.stderr).text();
	if (error) {
		console.error(`Error: ${error}`);
	} else if (stdout) {
		console.log(stdout);
	}

	callback();
};

const processFilesInBatches = (files) => {
	if (files.length === 0) {
		return;
	}

	const batch = files.slice(0, batchSize);
	const remaining = files.slice(batchSize);

	runBatch(batch, () => processFilesInBatches(remaining));
};

processFilesInBatches(files);
