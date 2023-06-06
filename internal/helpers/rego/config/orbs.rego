package circleci.config

#import data.circleci.utils

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

orbs_allow_types(allowed_orb_types) = { error |
#    known_orb_types = {"certified", "partner", "public"}
#    #invalid_orb_types = utils.to_set(allowed_orb_types) - known_orb_types
#    invalid_orb_types = allowed_orb_types - known_orb_types
#    count(invalid_orb_types) > 0
#    error := sprintf("invalid orb type(s) given in orbs_allow_types: %s", [concat(", ", invalid_orb_types)])

    disallowed_orbs = validate_orb_types(allowed_orb_types, orbs)
    count(disallowed_orbs) > 0
    error = sprintf("orb types are not allowed for these orbs: %s", [disallowed_orbs])
}
