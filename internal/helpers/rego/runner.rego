package circleci.config

resource_class_by_project(reservered_classes_to_projects) = {reason | 
        some name
        class = input.jobs[name].resource_class
        allowed_projects = reservered_classes_to_projects[class]
        not allowed_projects[data.meta.project_id]
        reason := sprintf("project is not allowed to use resource_class %q declared in job %q", [class, name])
}
