import type { Command } from "commander";
import { resolveContext } from "../cli.ts";
import { printOutput } from "../output.ts";

export function registerJobCommands(program: Command): void {
  program
    .command("list-jobs")
    .description("List all jobs")
    .action(async () => {
      const ctx = await resolveContext(program);
      const jobs = await listJobsRecursive(ctx.client, "");
      if (ctx.json) {
        printOutput(jobs, true);
      } else {
        for (const job of jobs) {
          console.log(job.fullName ?? job.name);
        }
      }
    });

  program
    .command("get-job")
    .description("Get job configuration XML")
    .argument("<name>", "job name (use folder/name for nested jobs)")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await ctx.client.getText(`${ctx.client.jobUrl(name)}/config.xml`);
      process.stdout.write(xml);
    });

  program
    .command("create-job")
    .description("Create a new job from XML on stdin")
    .argument("<name>", "job name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await ctx.client.readStdin();
      const parts = name.split("/");
      const jobName = parts.pop()!;
      const folderPath = parts.length > 0 ? ctx.client.jobUrl(parts.join("/")) : "";
      await ctx.client.post(
        `${folderPath}/createItem?name=${encodeURIComponent(jobName)}`,
        { body: xml, contentType: "application/xml" },
      );
      console.log(`Job '${name}' created.`);
    });

  program
    .command("copy-job")
    .description("Copy an existing job")
    .argument("<src>", "source job name")
    .argument("<dest>", "destination job name")
    .action(async (src: string, dest: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(
        `/createItem?name=${encodeURIComponent(dest)}&mode=copy&from=${encodeURIComponent(src)}`,
        { contentType: "application/x-www-form-urlencoded" },
      );
      console.log(`Job '${src}' copied to '${dest}'.`);
    });

  program
    .command("delete-job")
    .description("Delete a job")
    .argument("<name>", "job name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(`${ctx.client.jobUrl(name)}/doDelete`);
      console.log(`Job '${name}' deleted.`);
    });

  program
    .command("update-job")
    .description("Update job configuration from XML on stdin")
    .argument("<name>", "job name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await ctx.client.readStdin();
      await ctx.client.post(`${ctx.client.jobUrl(name)}/config.xml`, {
        body: xml,
        contentType: "application/xml",
      });
      console.log(`Job '${name}' updated.`);
    });

  program
    .command("reload-job")
    .description("Reload job configuration from disk")
    .argument("<name>", "job name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(`${ctx.client.jobUrl(name)}/reload`);
      console.log(`Job '${name}' reloaded.`);
    });

  program
    .command("build")
    .description("Trigger a build")
    .argument("<name>", "job name")
    .option("-p <params...>", "build parameters (key=value)")
    .option("--follow", "follow build console output", false)
    .option("--wait", "wait for build to complete", false)
    .action(async (name: string, opts: { p?: string[]; follow: boolean; wait: boolean }) => {
      const ctx = await resolveContext(program);
      const jobUrl = ctx.client.jobUrl(name);

      let buildPath: string;
      let body: string | undefined;
      let contentType: string | undefined;

      if (opts.p && opts.p.length > 0) {
        buildPath = `${jobUrl}/buildWithParameters`;
        const params = new URLSearchParams();
        for (const param of opts.p) {
          const eq = param.indexOf("=");
          if (eq === -1) throw new Error(`Invalid parameter format: ${param}. Expected key=value`);
          params.append(param.slice(0, eq), param.slice(eq + 1));
        }
        body = params.toString();
        contentType = "application/x-www-form-urlencoded";
      } else {
        buildPath = `${jobUrl}/build`;
      }

      const res = await ctx.client.post(buildPath, { body, contentType, raw: true });
      if (!res.ok && res.status !== 201) {
        throw new Error(`Build trigger failed: ${res.status} ${res.statusText}`);
      }

      const queueUrl = res.headers.get("Location");
      if (!queueUrl) {
        console.log("Build triggered.");
        return;
      }

      if (!opts.follow && !opts.wait) {
        console.log("Build triggered.");
        return;
      }

      const buildNumber = await waitForBuildNumber(ctx.client, queueUrl);
      if (!buildNumber) {
        console.error("Could not determine build number.");
        process.exit(1);
      }

      if (opts.follow) {
        await followConsole(ctx.client, jobUrl, buildNumber);
      } else if (opts.wait) {
        await waitForCompletion(ctx.client, jobUrl, buildNumber);
        console.log(`Build #${buildNumber} completed.`);
      }
    });
}

interface JenkinsJob {
  name: string;
  fullName?: string;
  url: string;
  color?: string;
  _class: string;
  jobs?: JenkinsJob[];
}

async function listJobsRecursive(
  client: import("../client.ts").JenkinsClient,
  path: string,
): Promise<JenkinsJob[]> {
  const prefix = path ? `${client.jobUrl(path)}` : "";
  const data = await client.getJson<{ jobs: JenkinsJob[] }>(
    `${prefix}/api/json?tree=jobs[name,fullName,url,color,_class,jobs[name,fullName,url,color,_class]]`,
  );
  const result: JenkinsJob[] = [];
  for (const job of data.jobs ?? []) {
    if (job._class?.includes("Folder") || job._class?.includes("OrganizationFolder")) {
      const childPath = path ? `${path}/${job.name}` : job.name;
      const children = await listJobsRecursive(client, childPath);
      result.push(...children);
    } else {
      result.push(job);
    }
  }
  return result;
}

async function waitForBuildNumber(
  client: import("../client.ts").JenkinsClient,
  queueUrl: string,
): Promise<number | null> {
  const queueApiUrl = queueUrl.endsWith("/")
    ? `${queueUrl}api/json`
    : `${queueUrl}/api/json`;

  const path = new URL(queueApiUrl).pathname;

  for (let i = 0; i < 120; i++) {
    await Bun.sleep(1000);
    try {
      const data = await client.getJson<{
        executable?: { number: number };
        cancelled?: boolean;
      }>(path);
      if (data.cancelled) return null;
      if (data.executable?.number) return data.executable.number;
    } catch {
      // queue item may not be available yet
    }
  }
  return null;
}

async function followConsole(
  client: import("../client.ts").JenkinsClient,
  jobUrl: string,
  buildNumber: number,
): Promise<void> {
  let offset = 0;
  while (true) {
    const res = await client.get(
      `${jobUrl}/${buildNumber}/logText/progressiveText?start=${offset}`,
      { raw: true },
    );
    const text = await res.text();
    if (text) {
      process.stdout.write(text);
    }
    const newOffset = res.headers.get("X-Text-Size");
    if (newOffset) {
      offset = parseInt(newOffset, 10);
    }
    const moreData = res.headers.get("X-More-Data");
    if (moreData !== "true") break;
    await Bun.sleep(1000);
  }
}

async function waitForCompletion(
  client: import("../client.ts").JenkinsClient,
  jobUrl: string,
  buildNumber: number,
): Promise<void> {
  for (let i = 0; i < 3600; i++) {
    const data = await client.getJson<{ building: boolean; result: string | null }>(
      `${jobUrl}/${buildNumber}/api/json?tree=building,result`,
    );
    if (!data.building) return;
    await Bun.sleep(2000);
  }
  throw new Error("Build did not complete within timeout.");
}
