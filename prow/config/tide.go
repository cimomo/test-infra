/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/test-infra/prow/github"
)

// Tide is config for the tide pool.
type Tide struct {
	// SyncPeriodString compiles into SyncPeriod at load time.
	SyncPeriodString string `json:"sync_period,omitempty"`
	// SyncPeriod specifies how often Tide will sync jobs with Github. Defaults to 1m.
	SyncPeriod time.Duration `json:"-"`
	// Queries must not overlap. It must be impossible for any two queries to
	// ever return the same PR.
	// TODO: This will only be possible when we allow specifying orgs. At that
	//       point, verify the above condition.
	Queries []TideQuery `json:"queries,omitempty"`

	// A key/value pair of an org/repo as the key and merge method to override
	// the default method of merge. Valid options are squash, rebase, and merge.
	MergeType map[string]github.PullRequestMergeType `json:"merge_method,omitempty"`

	// URL for tide status contexts.
	// We can consider allowing this to be set separately for separate repos, or
	// allowing it to be a template.
	TargetURL string `json:"target_url,omitempty"`

	// MaxGoroutines is the maximum number of goroutines spawned inside the
	// controller to handle org/repo:branch pools. Defaults to 20. Needs to be a
	// positive number.
	MaxGoroutines int `json:"max_goroutines,omitempty"`
}

// MergeMethod returns the merge method to use for a repo. The default of merge is
// returned when not overridden.
func (t *Tide) MergeMethod(org, repo string) github.PullRequestMergeType {
	name := org + "/" + repo

	v, ok := t.MergeType[name]
	if !ok {
		if ov, found := t.MergeType[org]; found {
			return ov
		}

		return github.MergeMerge
	}

	return v
}

// TideQuery is turned into a GitHub search query. See the docs for details:
// https://help.github.com/articles/searching-issues-and-pull-requests/
// If we choose to add orgs or branches then be sure to update the logic
// for listing all PRs in the tide package.
type TideQuery struct {
	Repos []string `json:"repos,omitempty"`

	Labels        []string `json:"labels,omitempty"`
	MissingLabels []string `json:"missingLabels,omitempty"`

	ReviewApprovedRequired bool `json:"reviewApprovedRequired,omitempty"`
}

func (tq *TideQuery) Query() string {
	toks := []string{"is:pr", "state:open"}
	for _, r := range tq.Repos {
		toks = append(toks, fmt.Sprintf("repo:\"%s\"", r))
	}
	for _, l := range tq.Labels {
		toks = append(toks, fmt.Sprintf("label:\"%s\"", l))
	}
	for _, l := range tq.MissingLabels {
		toks = append(toks, fmt.Sprintf("-label:\"%s\"", l))
	}
	if tq.ReviewApprovedRequired {
		toks = append(toks, "review:approved")
	}
	return strings.Join(toks, " ")
}

// AllPRs returns all open PRs in the repos covered by the query.
func (tq *TideQuery) AllPRs() string {
	toks := []string{"is:pr", "state:open"}
	for _, r := range tq.Repos {
		toks = append(toks, fmt.Sprintf("repo:\"%s\"", r))
	}
	return strings.Join(toks, " ")
}
