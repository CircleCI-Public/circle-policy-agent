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


## `require_orbs`

This function requires a config to contain orbs based on the orb names. Versions should not 
be included in the provided list of orbs.

### Definition
```
require_orbs([string])
returns { string: string }
```

### Usage

```
package org
import future.keywords
import data.circleci.config
require_security_orbs = config.require_orbs(["circleci/security", "foo/bar"])
enable_rule["require_security_orbs"]
hard_fail["require_security_orbs"] {
	require_security_orbs
}
```

## `require_orbs_version`

This function requires a policy to contain orbs based on the orb name and version.

### Definition
```
require_orbs_version([string])
returns { string: string }
```

### Usage

```
package org
import future.keywords
import data.circleci.config
require_orbs_versioned = config.require_orbs_version(["circleci/security@1.2.3", "foo/bar@4.5.6"])
enable_rule["require_orbs_versioned"]
hard_fail["require_orbs_versioned"] {
	require_orbs_versioned
}
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