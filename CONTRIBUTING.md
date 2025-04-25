# 🙌 Contributing to Customer Sentiment Intelligence (CSI)

First off, thanks for taking the time to contribute! 🚀
Whether it's a bug report, new feature, or a fix for a typo, we appreciate every contribution.

---

## 🧱 Ground Rules

- **One feature/fix per pull request.** This keeps things focused and easier to review.
- **Respect the architecture.** If you're unsure, ask in the issue or open a draft PR for feedback.
- **Write tests.** If it's testable, test it.
- **Format your code.** Use `go fmt` for Go and relevant linters for other tools.
- **Follow semantic commits.** Example: `feat(router): add Slack integration for issue routing`

---

## 🛠️ How to Contribute

1. **Fork the repo** and clone it.
2. Run `make setup` (if available) or manually install dependencies.
3. Create a branch: `git checkout -b feat/my-new-feature`
4. Make your changes and **add tests** if needed.
5. Run linter and tests: `make lint && make test`
6. Commit with semantic message: `git commit -m "fix(scraper): handle Twitter rate limit edge case"`
7. Push to your fork and open a PR.

---

## 🧪 Testing Guide

- Use the built-in test framework (`testing` package in Go).
- For AI components in Python, use `pytest`.
- Mock external APIs to avoid unnecessary calls.
- Run `go test ./...` at the root level to verify everything works.

---

## 🤝 Code of Conduct

Be respectful, inclusive, and supportive.\
We follow the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/) as our Code of Conduct.

---

## 💬 Communication

- For issues/bugs, open a [GitHub Issue](../../issues)
- For feature discussions, open a draft PR or join our community Slack

---

## 📄 License

By contributing, you agree that your contributions will be licensed under the same license as the project: MIT.

