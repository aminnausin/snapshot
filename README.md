# [GitHub Statistics Snapshot](https://github.com/aminnausin/snapshot)

<!--
https://github.community/t/support-theme-context-for-images-in-light-vs-dark-mode/147981/84
-->
<a href="https://github.com/aminnausin/snapshot">
<img src="https://github.com/aminnausin/snapshot/blob/main/generated/overview.svg#gh-dark-mode-only" alt="snapshot overview image for dark mode"/>
<img src="https://github.com/aminnausin/snapshot/blob/main/generated/languages.svg#gh-dark-mode-only" alt="snapshot languages image for dark mode"/>
<img src="https://github.com/aminnausin/snapshot/blob/main/generated/overview.svg#gh-light-mode-only" alt="snapshot overview image for light mode"/>
<img src="https://github.com/aminnausin/snapshot/blob/main/generated/languages.svg#gh-light-mode-only" alt="snapshot languages image for light mode"/>
</a>

Generate visualisations of GitHub user and repository statistics with GitHub
Actions. Visualisations can include data for both private repositories, and for
repositories you have contributed to, but do not own.

Generated images automatically switch between GitHub light theme and GitHub
dark theme.

## Background

This project is a Go-based reimplementation of [@jstrieb's](https://github.com/jstrieb) github-stats. I built this as a learning exercise and have recreated the all functionality using Go with a few additional features and fixes.

### New Features

- Optionally include profile view counts using [antonkomarev/github-profile-views-counter](https://github.com/antonkomarev/github-profile-views-counter)
- Better performance, reducing the number of GitHub Action minutes consumed every day

### Fixes

- Correctly differentiate between forked repos and contributions on open source repos
- Use GitHub GraphQL API to get lines changed instead of the REST API, overcoming the 202 Accepted Errors
- Make all configuration use environment variables / GitHub secrets instead of requiring edits to the workflow files
- Probably some others that I forgot about...

### Todo

- Parallelise requests and generation functions to increase speed even further

## Installation

1. Create a classic personal access token at [github.com/settings/tokens](https://github.com/settings/tokens) with the following permissions:

    - `read:user`
    - `repo`

    Copy the access token when it is generated – if you lose it, you will have to regenerate the token.

2. [Generate a new repository from this template](https://github.com/aminnausin/snapshot/generate).

3. Add the following secrets to your repository at [this link](../../settings/secrets/actions):

    - (Required) `ACCESS_TOKEN` — the token you generated earlier

4. Run the workflow in the [Actions tab](../../actions/workflows/main.yml?query=workflow%3A"Generate+Snapshot") (“Run workflow” button) to generate your stats for the first time.

The images will be automatically regenerated every 24 hours, but they can be regenerated manually by triggering the workflow.

To add your snapshot to your GitHub Profile README, copy and paste the following. Change the `username` value to your GitHub username.

``` md
![](https://raw.githubusercontent.com/username/snapshot/main/generated/overview.svg#gh-dark-mode-only)
![](https://raw.githubusercontent.com/username/snapshot/main/generated/overview.svg#gh-light-mode-only)
```

``` md
![](https://raw.githubusercontent.com/username/snapshot/main/generated/languages.svg#gh-dark-mode-only)
![](https://raw.githubusercontent.com/username/snapshot/main/generated/languages.svg#gh-light-mode-only)
```

## Configuration Options

You can add the following (optional) secrets to tweak the generated image:

- `EXCLUDED` — comma-separated list of repos to exclude (owner/name)

- `EXCLUDED_LANGS` — comma-separated list of languages to exclude from your snapshot. e.g., `html,tex,Jupyter Notebook`

- `INCLUDE_FORKED_REPOS` — set to `true` to include repositories you have forked (i.e., copies of someone else’s repo under your account). These are counted only if you are the owner of the forked repo.

- `INCLUDE_EXTERNAL_REPOS` — set to `true` to include repositories you’ve contributed to (e.g. via pull requests or reviews) but don’t own or have write access to, such as open source projects.

- `INCLUDE_PROFILE_VIEWS` — set to `true` if you're using [antonkomarev/github-profile-views-counter](https://github.com/antonkomarev/github-profile-views-counter)

## Support the Project

There are a few things you can do to support the project:

- Star the repository (and follow me on GitHub for more)
- Share and upvote on sites like Twitter, Reddit, and Hacker News
- Report any bugs, glitches, or errors that you find
- Link back to this repository so that others can generate their own snapshot images

## Related Projects

- Original version [jstrieb/github-stats](https://github.com/jstrieb/github-stats)
- Makes use of [GitHub Octicons](https://primer.style/octicons/) to precisely match the GitHub UI
