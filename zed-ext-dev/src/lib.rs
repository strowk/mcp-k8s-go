use std::{env, fs, path::Path};
use zed_extension_api::{self as zed, latest_github_release, Command};

const PACKAGE_NAME: &str = "@strowk/mcp-k8s";
const PACKAGE_VERSION: &str = "0.0.18";
const SERVER_PATH: &str = "node_modules/@strowk/mcp-k8s/bin/cli";

struct K8sContextServerExtension {}

impl zed::Extension for K8sContextServerExtension {
    fn new() -> Self
    where
        Self: Sized,
    {
        K8sContextServerExtension {}
    }

    fn context_server_command(
        &mut self,
        context_server_id: &zed_extension_api::ContextServerId,
        project: &zed::Project,
    ) -> zed_extension_api::Result<zed::Command> {
        Ok(Command {
            command: get_path_to_context_server_executable()?,
            args: get_args_for_context_server()?,
            env: get_env_for_context_server()?,
        })
    }
}

fn get_path_to_context_server_executable() -> zed_extension_api::Result<String> {
    // Ok("go".to_string())
    // Ok("wgo".to_string())
    Ok("arelo".to_string())
}

fn get_args_for_context_server() -> zed_extension_api::Result<Vec<String>> {
    // Ok(vec!["run".to_string(), "C:/work/mcp-k8s-go/main.go".to_string()])
    // Ok(vec!["run".to_string(), "-stdin".to_string(), "-cd".to_string(), "C:/work/mcp-k8s-go".to_string(), "main.go".to_string()])
    Ok(vec![
        "-p".to_string(),
        "**/*.go".to_string(),
        "-i".to_string(),
        "**/.*".to_string(),
        "-i".to_string(),
        "**/*_test.go".to_string(),
        "-t".to_string(),
        "C:/work/mcp-k8s-go".to_string(),
        "--".to_string(),
        "mcptee".to_string(),
        "C:/work/mcp-k8s-go/dev.log.yaml".to_string(),
        // "list_k8s_namespaces_prompt.exe".to_string(),
        "go".to_string(),
        "run".to_string(),
        "-C".to_string(),
        "C:/work/mcp-k8s-go".to_string(),
        "main.go".to_string(),
    ])
}

fn get_env_for_context_server() -> zed_extension_api::Result<Vec<(String, String)>> {
    Ok(vec![])
}

zed::register_extension!(K8sContextServerExtension);
