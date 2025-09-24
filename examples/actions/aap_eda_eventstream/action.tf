terraform {
  required_providers {
    aap = {
      source = "ansible/aap"
    }
  }
}

provider "aap" {
  host     = "https://AAP_HOST"
  username = "ansible"
  password = "test123!"
}

resource "terraform_data" "trigger" {
  input = "%s"
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aap_eda_eventstream.create]
    }
  }
}

action "aap_eda_eventstream" "create" {
  config {
    limit             = "infra"
    template_type     = "job"
    job_template_name = "After Create Job Template"
    organization_name = "Default"
    event_stream_config = {
      username = "event-stream-username"
      password = "event-stream-password"
      url      = "http://aap-event-stream-url.example.com/"
    }
  }
}
