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
        let version = zed::npm_package_installed_version(PACKAGE_NAME)?;
        if version.as_deref() != Some(PACKAGE_VERSION) {
            zed::npm_install_package(PACKAGE_NAME, PACKAGE_VERSION)?;
        }

        let release = latest_github_release(
            "strowk/mcp-k8s-go",
            zed_extension_api::GithubReleaseOptions {
                pre_release: false,
                require_assets: true,
            },
        )?;

        let (platform, arch) = zed::current_platform();
        let asset_name = format!(
            "mcp-k8s-go_{os}_{arch}.{format}",
            arch = match arch {
                zed::Architecture::Aarch64 => "arm64",
                zed::Architecture::X86 => "i386",
                zed::Architecture::X8664 => "x86_64",
            },
            os = match platform {
                zed::Os::Mac => "Darwin",
                zed::Os::Linux => "Linux",
                zed::Os::Windows => "Windows",
            },
            format = match platform {
                zed::Os::Mac => "tar.gz",
                zed::Os::Linux => "tar.gz",
                zed::Os::Windows => "zip",
            }
        );

        let asset = release
            .assets
            .iter()
            .find(|asset| asset.name == asset_name)
            .ok_or_else(|| {
                format!(
                    "could not find asset {:?} in {:?}",
                    asset_name, release.assets
                )
            })?;

        let directory = format!("mcp-k8s-go-{}", release.version);

        fs::create_dir_all(&directory)
            .map_err(|error| format!("could not create directory {directory} due to '{error}'"))?;

        let asset_path = format!("{}/{}", directory, asset_name);

        let dir_exists = fs::metadata(&asset_path).is_ok();

        let archive_type = match platform {
            zed::Os::Mac => zed::DownloadedFileType::GzipTar,
            zed::Os::Linux => zed::DownloadedFileType::GzipTar,
            zed::Os::Windows => zed::DownloadedFileType::Zip,
        };

        let bin_name = match platform {
            zed::Os::Windows => "mcp-k8s-go.exe",
            _ => "mcp-k8s-go",
        };
        let bin_path = format!("{asset_path}/{bin_name}");

        let current_dir = env::current_dir()
            .map_err(|err| format!("could not resolve current directory: {err}"))?
            .display()
            .to_string();

        // TODO: this looks like a hacky workaround, maybe it is needed due
        // to something not quite implemented on wasm side, but maybe this could be fixed?
        //
        // Windows is somewhat peculliar in that it uses backslashes for paths
        // and canonicalization does not work, as well as "join" later
        // does not continue building path in system specific way, hence we try
        // to transform windows-like path to sane path here
        let sanitized_current_dir = match platform {
            zed::Os::Windows => current_dir.replace("\\", "/").replace("/C:/", "C:/"),
            _ => current_dir,
        };

        let execute_path = Path::new(&sanitized_current_dir)
            .join(&directory)
            .join(&asset_name)
            .join(bin_name);

        let command = execute_path
            .to_str()
            .ok_or_else(|| format!("could not convert path to string: {:?}", execute_path))?
            .to_string();

        if !dir_exists {
            zed::download_file(&asset.download_url, &asset_path, archive_type)
                .map_err(|err| format!("could not download file: {err}"))?;
            zed::make_file_executable(&bin_path)
                .map_err(|err| format!("could not make file executable: {err}"))?;
        }

        let entries = fs::read_dir(sanitized_current_dir)
            .map_err(|e| format!("failed to list working directory {e}"))?;
        for entry in entries {
            let entry = entry.map_err(|e| format!("failed to load directory entry {e}"))?;
            if entry.file_name().to_str() != Some(&directory) {
                fs::remove_dir_all(entry.path()).ok();
            }
        }

        Ok(Command {
            command: command,
            args: vec![],
            env: vec![],
        })
    }
}

fn get_path_to_context_server_executable() -> zed_extension_api::Result<String> {
    Ok("go".to_string())
}

fn get_args_for_context_server() -> zed_extension_api::Result<Vec<String>> {
    Ok(vec!["run".to_string(), "../main.go".to_string()])
}

fn get_env_for_context_server() -> zed_extension_api::Result<Vec<(String, String)>> {
    Ok(vec![])
}

zed::register_extension!(K8sContextServerExtension);
