# Fork maintenance instructions

Before changing or synchronizing this repository, read `CUSTOM_FEATURES.md` in full.
It records the current custom behavior and features deliberately removed by the
maintainer. Do not infer current requirements from older commits alone.

- Merge upstream updates into this fork; never replace the whole tree with an
  upstream checkout or discard unrelated customizations when resolving conflicts.
- Preserve each active customization unless the human explicitly retires it.
  Update the inventory, implementation references and tests in the same change.
- Run regression tests for affected customizations as well as upstream changes.
  Report failures honestly; never fabricate prices or disable billing to pass tests.
- Build and test Go in server Docker. Do not install a local Go environment.
- On the 8 GiB production host, serialize Go builds/tests in a container with
  an explicit hard memory limit (at most 4 GiB), CPU limit and `go -p 1`.
  Never run multiple unbounded compilers on the production host. Keep production
  headroom; a killed test compiler is not a passing test. Reuse caches to avoid
  repeated cold builds, but do not skip a failed regression to ship a release.
- Verification must not call real model accounts or spend quota. Use synthetic
  mocks and isolated test databases; never use production DB credentials for tests.
- Production deployment requires user authorization, a recoverable code/database
  backup, staged verification, health checks and an exact rollback target.
- Never commit credentials, OAuth tokens, deployment secrets or database dumps.
- Before handoff, report the upstream tag, Git commit, production version and
  verification coverage separately. A successful build is not proof of deployment.

Historical documents such as `CODEX_QUOTA_OVERDRAFT_CUSTOMIZATION.md` describe
older baselines. For current defaults and retired features, use
`CUSTOM_FEATURES.md` and executable code/tests.
