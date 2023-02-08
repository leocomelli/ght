# ght

ght (GitHub Template) helps us create or maintain a repository according to some standard settings. We must use a JSON file to configure the repository settings, the JSON format is the same as that used by the [GitHub REST API](https://docs.github.com/en/rest?apiVersion=2022-11-28).

The supported settings are:

1. [Create an organization repository](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#create-an-organization-repository);
2. [Create a repository for the authenticated user](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#create-a-repository-for-the-authenticated-user);
3. [Create a repository using a template](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#create-a-repository-using-a-template);
4. [Update a repository](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#update-a-repository);
5. [Replace all repository topics](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#replace-all-repository-topics);
6. [Update branch protection](https://docs.github.com/en/rest/branches/branch-protection?apiVersion=2022-11-28#update-branch-protection);
7. [Create commit signature protection](https://docs.github.com/en/rest/branches/branch-protection?apiVersion=2022-11-28#create-commit-signature-protection);
8. [Create a PR template](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/creating-a-pull-request-template-for-your-repository);
9. [Create an Issue template](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/manually-creating-a-single-issue-template-for-your-repository);

## CLI flags

There are some parameters that must be provided as CLI flags:

```text
  -b, --branches strings     the names of the branches to which the protection rules will be applied
  -v, --debug                enable debug mode
  -d, --description string   a short description of the repository
  -h, --help                 help for repo
  -n, --name string          the name of the repository
  -o, --owner string         the name of the owner, can be an organization or an authenticated user
  -t, --template string      the name of the JSON file that contains the template, can be a local or remote file
  -l, --topics strings       an array of topics to add to the repository
```

## Usage

The example of usage below uses all available flags, but only `owner`, `name` and `template` are required.

```bash
ght repo --owner leocomelli \
         --name ght \
         --description create or maintain a repository according to some standard options \
         --topics github,golang,go \
         --branches main \
         --template example.json \
         --debug
```

Using this template file, a new repository will be created (or updated if it already exists), and the repository settings will be defined according to the `repository` node. In addition, a set of branch protection rules will be created following the `branch_protection` and `required_signed_commits` nodes. Note that if it is a new repository and we are using the default branch(`main`), the `"auto_init": true` must be used on the `repository` node.

```json
{
  "repository": {
    "private": true,
    "has_issues": false,
    "has_projects": false,
    "has_wiki": false,
    "allow_squash_merge": true,
    "allow_merge_commit": false,
    "allow_rebase_merge": false,
    "delete_branch_on_merge": true,
    "squash_merge_commit_title": "PR_TITLE",
    "squash_merge_commit_message": "COMMIT_MESSAGES",
    "auto_init": true
  },
  "branch_protection": {
    "required_status_checks": {
      "strict": true,
      "checks": []
    },
    "required_pull_request_reviews": {
      "dismiss_stale_reviews": true,
      "require_code_owner_reviews": true
    },
    "enforce_admins": true
  },
  "required_signed_commits": true,
  "pull_request_template": "https://raw.githubusercontent.com/leocomelli/ght/main/examples/pr_template.md"
  "issue_template": "https://raw.githubusercontent.com/leocomelli/ght/main/examples/issue_template.md"
}
```

The `pull_request_template` could be a local or remote file, as well as the `issue_template`.

## ght _vs_ GitHub feature (create from a template)

The ght ensures that some settings will be applied when a repository is created or updated, whereas the GitHub feature is similar to forking a repository. In general, the ght is about settings and the GitHub feature is about branches and directory structure.
To get more information, read the documentation: [Creating a repository from a template
](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-repository-from-a-template).

Don't forget we can use ght and GitHub features together (check [here](https://github.com/leocomelli/ght/blob/main/testing/simple-repo-template.json)).
