# Prompt log

Every prompt I sent to the AI assistant while building this project, in order and word for word. Each entry ends with one line on what it produced.

---

## 1. Ground rules and requirements

```
We're building a take-home for a payments company: a full-stack calculator. The backend is a Go REST microservice, and the frontend is React with TypeScript. The frontend calls the backend for every calculation and never does the maths itself.

Rules for the whole build:
1. We go one step at a time. Build only what each prompt asks for, then stop. Show me the files changed and why, and run whatever proves it works.
2. Log every prompt I send for this project in docs/PROMPTS.md, word for word, numbered, each with one line on what it produced.
3. Keep docs/REQUIREMENTS.md as a checklist of every requirement in the assignment. Tick an item only when it's done and proven.
4. Every approved step is one Conventional Commit. No AI co-author lines and no em dashes, anywhere.
5. Correctness and clarity beat features. If something's out of scope, put it under "next steps" instead of building it.

Start by creating docs/REQUIREMENTS.md from the assignment text I'll paste below, and docs/PROMPTS.md with this prompt as entry 1.

Build a full-stack calculator application with a React frontend and a backend microservice. The frontend should consume the backend API to perform basic and advanced arithmetic operations. Focus on clean design, maintainable code, and testable architecture.Requirements
Functional
Operations:

* Addition, Subtraction, Multiplication, Division
* Optional: Exponentiation, Square Root, Percentage

Frontend (React):

* Intuitive UI for entering input and displaying results
* Input validation and error handling
* Responsive design (basic mobile support)

Backend (REST API):

* Expose endpoints for calculator operations
* Validate input and handle edge cases (division by zero, invalid data)
* Return results in JSON format

Non-Functional

* Clean, readable, and idiomatic code (frontend and backend)
* Unit tests covering key functionality for both layers
* Documentation: setup instructions, API usage, and design rationale
* Optional: Dockerfile for full-stack deployment

Constraints

* Frontend: React (TypeScript preferred)
* Backend: Go is perferred

Deliverables

1. Git repository with frontend and backend code
2. README with setup instructions, API examples, and design decisions
3. Unit tests and coverage report
4. Optional: Dockerfile to run frontend + backend together

Instructions

1. Use any AI tooling you would like
2. Spend ~2–4 hours on this assignment. Prioritize correctness, clarity, and maintainability over extra features.
3. Push your solution to GitHub, GitLab, or another Git repository.
4. Share the repository link with us for evaluation.
5. Share any prompts that you used in your work
6. Make sure your README includes:
   * Setup instructions
   * How to run the frontend and backend
   * Examples of API calls (if using REST)
   * Design decisions or assumptions
```

Produced: `docs/REQUIREMENTS.md`, a 41-item checklist with IDs and a proof column, and this prompt log.

---

## 2. Repo skeleton

```
Create the repo skeleton, with no app logic yet:
- backend/ for the Go service, and frontend/ for the React app
- docs/ (already started)
- A root Makefile that will delegate to backend and frontend. For now it only has a help target that lists the targets.
- A .gitignore covering Go builds, coverage output, generated mocks, node_modules, dist, .env files and OS files
- An .editorconfig: LF line endings, a final newline, tabs for Go, 2 spaces for TS, JSON, YAML and Markdown
- An MIT LICENSE in my name
- A minimal README stub with the project name and one line describing it

Show me the tree when you're done.
```

Produced: `backend/` and `frontend/` placeholders, a self-documenting root `Makefile` with `help`, `.gitignore`, `.editorconfig`, MIT `LICENSE` and a README stub.

---

## 3. CI pipeline

```
Add a GitHub Actions workflow at .github/workflows/ci.yml. It runs on pushes and pull requests to main and has two jobs.
- backend: set up Go from the version in go.mod, cache modules, generate mocks, run golangci-lint, run the tests with the race detector and coverage, then build.
- frontend: set up Node LTS with an npm cache, npm ci, then type-check, lint, run the tests with coverage, and build.
Both jobs should fail loudly on any warning that matters. Upload both coverage reports as artifacts. Use pinned action versions. The jobs can't pass yet; that's expected. Explain each step in one line.
```

Produced: `.github/workflows/ci.yml` with backend and frontend jobs, actions pinned to commit SHAs, golangci-lint pinned, zero-warning lint, and coverage artifacts. Validated with actionlint.
