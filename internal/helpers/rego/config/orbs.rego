package circleci.config

import data.circleci.utils
import future.keywords.in

orbs[name] = version {
    some orb in input["orbs"]
    [name, version] := split(orb, "@")
}

parameterized_orbs_in_input(orbs) = { orb: msg |
some orb in orbs
utils.is_parameterized_expression(orb)
msg := sprintf("invalid orb: %s - parameterized orbs are disallowed", [orb])
}

ban_orbs(orb_names) = object.union({ orb_name: msg | 
  orb_name := orb_names[_]
  orbs[orb_name]
  msg := sprintf("%s orb is not allowed in CircleCI configuration", [orb_name])
},parameterized_orbs_in_input(input.orbs))

ban_orbs_version(banned_orbs) = object.union({ orb: msg | orb := banned_orbs[_]
  [name, version] := split(orb, "@")
  orbs[name] == version
  msg := sprintf("%s orb is not allowed in CircleCI configuration", [orb])
},parameterized_orbs_in_input(input.orbs))

orbs_allowlist(allowed_orbs) = { orb: msg | orb := input["orbs"][_]
  [name, _] := split(orb, "@")
  not allowed_orbs[name]
  msg := sprintf("%s orb is not allowed in CircleCI configuration", [name])
}
