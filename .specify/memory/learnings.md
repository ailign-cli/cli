# Learnings

Project-specific conventions and patterns discovered through code
reviews and development. Curated — not a changelog.

## Documentation

- Use generic/fictional examples in documentation (e.g.,
  `acme/web-api#123`) instead of real repository names, even
  internal ones. Avoids implying affiliation, leaking private repo
  names, and confusing access controls.

## PR hygiene

- When a PR accumulates features beyond the original description,
  update the PR body to list all changes. Reviewers (human and bot)
  judge scope against the description.
