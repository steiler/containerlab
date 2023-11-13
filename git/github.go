package git

import (
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	neturl "net/url"
	"strings"
)

// NewGitHubRepoFromURL parses the given url and returns a GitHubRepo.
func NewGitHubRepoFromURL(url *neturl.URL) (*GitHubRepo, error) {
	r := &GitHubRepo{
		GitRepoStruct: GitRepoStruct{
			URL: url,
		}}

	// trimming the leading and trailing slashes
	// so that splitPath will have the slashes between the elements only
	splitPath := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	// path needs to hold at least 2 elements,
	// user / org and repo
	if len(splitPath) < 2 || splitPath[0] == "" || splitPath[1] == "" {
		return nil, fmt.Errorf("%w %s", errInvalidURL, r.URL.String())
	}

	// github.dev links can be cloned using github.com
	if r.URL.Host == "github.dev" {
		r.URL.Host = "github.com"
	}

	// set CloneURL to the copy of the original URL
	// this is overriden later for repo urls with
	// paths containing blob or tree elements
	r.CloneURL = &neturl.URL{}
	*r.CloneURL = *r.URL

	// remove trailing slash from the path
	// as it bears no meaning for the clone url
	r.CloneURL.Path = strings.TrimSuffix(r.CloneURL.Path, "/")

	r.ProjectOwner = splitPath[0]

	// in case repo url has a trailing .git suffix, trim it
	r.RepositoryName = strings.TrimSuffix(splitPath[1], ".git")

	switch {
	case len(splitPath) == 2:
		return r, nil
	case len(splitPath) < 4:
		return nil, fmt.Errorf("%w invalid github path. should have either 2 or >= 4 path elements", errInvalidURL)
	}

	r.GitBranch = splitPath[3]

	switch {
	// path points to a file at a specific git ref
	case splitPath[2] == "blob":
		if !(strings.HasSuffix(r.URL.Path, ".yml") || strings.HasSuffix(r.URL.Path, ".yaml")) {
			return nil, fmt.Errorf("%w: topology file must have yml or yaml extension", errInvalidURL)
		}

		if len(splitPath)-1 > 4 {
			r.Path = splitPath[4 : len(splitPath)-1]
		}

		// overriding CloneURL Path to point to the git repo
		r.CloneURL.Path = "/" + splitPath[0] + "/" + splitPath[1]

		r.FileName = splitPath[len(splitPath)-1]

	// path points to a git ref (branch or tag)
	case splitPath[2] == "tree":
		if len(splitPath) > 4 {
			r.Path = splitPath[4:]
		}

		// overriding CloneURL Path to point to the git repo
		r.CloneURL.Path = "/" + splitPath[0] + "/" + splitPath[1]

		r.FileName = "" // no filename, a dir is referenced
	}

	return r, nil
}

// IsGitHubURL checks if the url is a github url.
func IsGitHubURL(url *neturl.URL) bool {
	return strings.Contains(url.Host, "github.com") ||
		strings.Contains(url.Host, "github.dev")
}

// GitHubRepo struct holds the parsed github url.
type GitHubRepo struct {
	GitRepoStruct
}

// IsGitHubShortURL returns true for github-friendly short urls
// such as srl-labs/containerlab.
func IsGitHubShortURL(s string) bool {
	split := strings.Split(s, "/")

	// only 2 elements are allowed
	if len(split) != 2 {
		return false
	}

	// dot is not allowed in the project owner
	if strings.Contains(split[0], ".") {
		return false
	}

	return true
}

func ExtractGitURLFromShort(user, repo, id string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s", user, repo, id)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	switch response.StatusCode {
	case 200:
		// all good, simply continue
	case 404:
		return "", fmt.Errorf("unable to retrieve pull request \"%s/%s#%s\". (%s) with status code %d This is probably not referencing a pull request", user, repo, id, url, response.StatusCode)
	default:
		return "", fmt.Errorf("unable to retrieve pull request \"%s/%s#%s\" (%s) with status code %d", user, repo, id, url, response.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// unmarshall response into the partial json struct
	var pullUrl partialGithubApiPullResponse
	err = json.Unmarshal(body, &pullUrl)
	if err != nil {
		return "", err
	}
	// check if the retrieved data was valid, if all the
	// expected fields are set.
	if err = pullUrl.Valid(); err != nil {
		return "", err
	}
	// return the URL of the pull request referenced branch
	return pullUrl.GetUrl(), nil
}

// partialGithubApiPullResponse partial struct for github pulls api
type partialGithubApiPullResponse struct {
	Head struct {
		Ref  string `json:"ref"`
		Repo struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			}
		} `json:"repo"`
	} `json:"head"`
}

// GetUrl composes the GitUrl from the partialGithubApiPullResponse
func (p *partialGithubApiPullResponse) GetUrl() string {
	return fmt.Sprintf("https://github.com/%s/%s/tree/%s", p.Head.Repo.Owner.Login, p.Head.Repo.Name, p.Head.Ref)
}

// Valid
func (p *partialGithubApiPullResponse) Valid() error {
	if p.Head.Ref == "" || p.Head.Repo.Name == "" || p.Head.Repo.Owner.Login == "" {
		return fmt.Errorf("unable to determine branch")
	}
	return nil
}
