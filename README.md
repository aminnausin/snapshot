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

<img src="./templates/overview.svg" alt="snapshot overview image for light mode"/>
<img src="./templates/languages.svg" alt="snapshot languages image for light mode"/>

Generate visualizations of GitHub user and repository statistics with GitHub
Actions. Visualizations can include data for both private repositories, and for
repositories you have contributed to, but do not own.

Generated images automatically switch between GitHub light theme and GitHub
dark theme.

## Background

This is a recreation of the same project created by [jstrieb](https://github.com/jstrieb) but written in GO.
I created this project to learn GO and make something with a known end goal. It behaves the same way as the original.

## Installation

1. Create a classic personal access token (not the default GitHub Actions token) in your [Github developer settings page](https://github.com/settings/tokens).
   Personal access token must have permissions: `read:user` and `repo`. Copy
   the access token when it is generated – if you lose it, you will have to
   regenerate the token.
2. [Create a copy of this repository.](https://github.com/aminnausin/snapshot/generate) Note: this is
   **not** the same as forking a copy because it copies everything fresh,
   without the huge commit history.
3. Go to the "Secrets" page of your copy of the repository. If this is the
   README of your copy, click [this link](../../settings/secrets/actions) to go
   to the "Secrets" page. Otherwise, go to the "Settings" tab of the
   newly-created repository and go to the "Secrets" page (bottom left).
4. Create a new secret with the name `ACCESS_TOKEN` and paste the copied
   personal access token as the value.
5. It is possible to change the type of statistics reported by adding other
   repository secrets.
    - To ignore certain repos, add them (in owner/name format e.g.,
      `jstrieb/github-stats`) separated by commas to a new secret—created as
      before—called `EXCLUDED`.
    - To ignore certain languages, add them (separated by commas) to a new
      secret called `EXCLUDED_LANGS`. For example, to exclude HTML, TeX and Jupyter Notebook you
      could set the value to `html,tex,Jupyter Notebook`.
    - To include statistics for forked repositories with
      contributions, add another secret called `EXCLUDE_FORKED_REPOS` with a value of `false`.
    - If you are using [antonkomarev/github-profile-views-counter](https://github.com/antonkomarev/github-profile-views-counter) on your profile readme,
      you can set a secret called `INCLUDE_PROFILE_VIEWS` with a value of `true` to add it to your snapshot card. Ideally, you include the standard image in the invisible mode.
6. Go to the [Actions
   Page](../../actions/workflows/main.yml?query=workflow%3A"Generate+Snapshot") and press "Run
   Workflow" on the right side of the screen to generate images for the first
   time.
    - The images will be automatically regenerated every 24 hours, but they can
      be regenerated manually by running the workflow this way.
7. Take a look at the images that have been created in the
   [`generated`](generated) folder.
8. To add your statistics to your GitHub Profile README, copy and paste the
   following lines of code into your markdown content. Change the `username`
   value to your GitHub username.

    ```md
    ![](https://raw.githubusercontent.com/username/snapshot/main/generated/overview.svg#gh-dark-mode-only)
    ![](https://raw.githubusercontent.com/username/snapshot/main/generated/overview.svg#gh-light-mode-only)
    ```

    ```md
    ![](https://raw.githubusercontent.com/username/snapshot/main/generated/languages.svg#gh-dark-mode-only)
    ![](https://raw.githubusercontent.com/username/snapshot/main/generated/languages.svg#gh-light-mode-only)
    ```

9. Link back to this repository so that others can generate their own
   statistics images.
10. Star this repo if you like it!

## Support the Project

There are a few things you can do to support the project:

-   Star the repository (and follow me on GitHub for more)
-   Share and upvote on sites like Twitter, Reddit, and Hacker News
-   Report any bugs, glitches, or errors that you find

## Related Projects

-   An iteration upon [jstrieb/github-stats](https://github.com/jstrieb/github-stats) using GO
-   Makes use of [GitHub Octicons](https://primer.style/octicons/) to precisely match the GitHub UI
