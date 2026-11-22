import { defineRailway, github, project, service } from "railway/iac";
// Railway infrastructure-as-code for Gamer's Paradise (ADR-009). This
// manages only the app service; Postgres and all variables stay in Railway.
//
//   gamers-paradise  deploys `main` -> https://gamers-paradise.up.railway.app
//
// checkSuites: Railway waits for the commit's GitHub checks (CI) to pass
// before deploying, so a red CI run never reaches the site.
//
// Plan/apply (see docs/development.md for the Windows note):
//   railway config plan  --environment production
//   railway config apply --environment production
export const partial = "gamers-paradise";

export default defineRailway(() =>
  project("gamers-paradise", {
    variables: { managed: false },
    resources: [
      service("gamers-paradise", {
        source: github("berryscottr/gamers-paradise", { branch: "main", checkSuites: true }),
        build: { buildEnvironment: "V3", builder: "DOCKERFILE", dockerfilePath: "Dockerfile" },
        // Restart policy "on failure" is Railway's default (stored as empty).
        deploy: { restartPolicyMaxRetries: 5 },
        healthcheck: "/healthz",
        healthcheckTimeout: 30,
        replicas: { "us-east4-eqdc4a": 1 },
      }),
    ],
  }),
);
