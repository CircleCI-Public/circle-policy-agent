# Circle CI Job Helpers

## `jobs`
`jobs` is a Rego object containing jobs that are present in the given CircleCI config file. It 
can be utilized by policies related to jobs.

### Definition
```
jobs = []string
```

Example `jobs` object:
```
[
    "job-a",
    "job-b",
    "job-c"
]
```

### Usage
```
package org
import future.keywords
import data.circleci.config

jobs := config.jobs
```


## `require_jobs`

This function requires a config to contain jobs based on the job names. Each job in the list of 
required jobs must be in at least one workflow within the config.

### Definition
```
require_jobs([string])
returns { string }
```

### Usage

```
package org
import future.keywords
import data.circleci.config
require_security_jobs = config.require_jobs(["security-check", "vulnerability-scan"])
enable_rule["require_security_jobs"]
hard_fail["require_security_jobs"] {
	require_security_jobs
}
```