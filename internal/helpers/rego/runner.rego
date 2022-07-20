package circleci.config

resource_class_by_project(reservered_classes_to_projects) = {reason | 
        some name
        class = input.jobs[name].resource_class
        not reservered_classes_to_projects[class][data.meta.project_id]
        reason := sprintf("project %q is not allowed to use resource_class %q declared in job %q", [data.meta.project_id, class, name])
}
