package circleci.config

import future.keywords.in

orbs[name] = version {
    some orb in input["orbs"]
    [name, version] := split(orb, "@")
}

ban_orbs(orb_names) = { orb_name: msg | orb_name := orb_names[_]
  orbs[orb_name]
  msg := sprintf("%s orb is not allowed in CircleCI configuration", [orb_name])
}

ban_orbs_version(banned_orbs) = { orb: msg | orb := banned_orbs[_]
  [name, version] := split(orb, "@")
  orbs[name] == version
  msg := sprintf("%s orb is not allowed in CircleCI configuration", [orb])
}

orbs_allowlist(allowed_orbs) = { orb: msg | orb := input["orbs"][_]
  [name, _] := split(orb, "@")
  not allowed_orbs[name]
  msg := sprintf("%s orb is not allowed in CircleCI configuration", [name])
}
