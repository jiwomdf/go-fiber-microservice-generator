import { defineConfig } from "vite";

const repoName = "go-fiber-microservice-generator";

export default defineConfig({
  base: process.env.GITHUB_ACTIONS ? `/${repoName}/` : "/",
});
