package circleci.config

import future.keywords.in

orbs[name] = version {
    some orb in input["orbs"]
    [name, version] := split(orb, "@")
}

require_orbs(orb_names) = { orb_name: msg | orb_name := orb_names[_]
  not orbs[orb_name]
  msg := sprintf("%s orb is required", [orb_name])
}

require_orbs_version(required_orbs) = { orb: msg | orb := required_orbs[_]
  [name, version] := split(orb, "@")
  not orbs[name] == version
  msg := sprintf("%s orb is required", [orb])
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