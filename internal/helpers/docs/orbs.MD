# Circle CI Orb Helpers

## `orbs`
`orbs` is a Rego object containing orbs and versions present in the given config file. It 
can be utilized by policies related to orbs.

### Definition
```
orbs[string] = string
```

Example `orbs` object:
```
{
    "circleci/security": "1.2.3",
    "circleci/foo": "3.2.1"
}
```

### Usage
```
package org
import future.keywords
import data.circleci.config

my_orbs := config.orbs
```


## `ban_orbs`

This function violates a policy if a config includes orbs based on the orb name. Versions should not 
be included in the provided list of orbs.

### Definition
```
ban_orbs_version([string])
returns { string: string }
```

### Usage

```
package org
import future.keywords
import data.circleci.config
ban_orbs = config.ban_orbs(["evilcorp/evil"])
enable_rule["ban_orbs"]
hard_fail["ban_orbs"] {
	ban_orbs
}
```

## `ban_orbs_version`

This function violates a policy if a config includes orbs based on the orb name and version.

### Definition
```
ban_orbs_version([string])
returns { string: string }
```

### Usage

```
package org
import future.keywords
import data.circleci.config
ban_orbs_versioned = config.ban_orbs_version(["evilcorp/evil@1.2.3", "foo/bar@4.5.6"])
enable_rule["ban_orbs_versioned"]
hard_fail["ban_orbs_versioned"] {
	ban_orbs_versioned
}
```