
use crate::{
    backups,
    config::Config,
    systemd,
    worlds,
};
use axum::{
    extract::State,
    http::{
        header,
        HeaderMap,
        HeaderValue,
        StatusCode,
    },
    response::{Html, IntoResponse, Response},
    routing::{get, post},
    Form,
    Router,
};
use base64::{
    engine::general_purpose::STANDARD,
    Engine,
};
use serde::Deserialize;
use std::sync::Arc;

#[derive(Clone)]
struct AppState {
    config: Arc<Config>,
}

#[derive(Deserialize)]
struct ActionForm {
    action: String,
    csrf_token: String,
}

fn authorized(headers: &HeaderMap, config: &Config) -> bool {
    let Some(value) = headers.get(header::AUTHORIZATION) else {
        return false;
    };

    let Ok(value) = value.to_str() else {
        return false;
    };

    let Some(encoded) = value.strip_prefix("Basic ") else {
        return false;
    };

    let Ok(decoded) = STANDARD.decode(encoded) else {
        return false;
    };

    let Ok(credentials) = String::from_utf8(decoded) else {
        return false;
    };

    let expected = format!(
        "{}:{}",
        config.web.username,
        config.web.password
    );

    credentials == expected
}

fn unauthorized() -> Response {
    let mut response = (
        StatusCode::UNAUTHORIZED,
        "Authentication required",
    ).into_response();

    response.headers_mut().insert(
        header::WWW_AUTHENTICATE,
        HeaderValue::from_static(
            "Basic realm=\"valheimctl\", charset=\"UTF-8\""
        ),
    );

    response
}

fn escape_html(input: &str) -> String {
    input
        .replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
        .replace('\'', "&#39;")
}

async fn dashboard(
    State(state): State<Arc<AppState>>,
    headers: HeaderMap,
) -> Response {
    if !authorized(&headers, &state.config) {
        return unauthorized();
    }

    let config = &state.config;

    let status = systemd::status(
        &config.server.systemd_unit
    );

    let (server_state, server_description) = match status {
        Ok(status) => (
            status.state,
            status.description,
        ),
        Err(error) => (
            "unknown".to_string(),
            format!("Status error: {error:#}"),
        ),
    };

    let entries = worlds::inspect(
        &config.worlds.directory
    );

    let (worlds_summary, worlds_list) = match entries {
        Ok(entries) => {
            let count = entries.len();

            let list = if entries.is_empty() {
                "<li>No entries found</li>".to_string()
            } else {
                entries.iter().map(|entry| {
                    format!(
                        "<li><code>{}</code> ({}) — {} files, {} bytes{}</li>",
                        escape_html(&entry.name),
                        escape_html(&entry.path.display().to_string()),
                        entry.files,
                        entry.bytes,
                        if entry.is_directory {
                            " (directory)"
                        } else {
                            ""
                        }
                    )
                }).collect::<String>()
            };

            (format!("{count} top-level entries"), list)
        }
        Err(error) => (
            format!("Could not inspect saves: {error:#}"),
            "<li>Save directory unavailable</li>".to_string(),
        ),
    };

    let template = include_str!("../templates/dashboard.html");

    let page = template
        .replace(
            "{{server_name}}",
            &escape_html(&config.server.name),
        )
        .replace(
            "{{server_state}}",
            &escape_html(&server_state),
        )
        .replace(
            "{{server_description}}",
            &escape_html(&server_description),
        )
        .replace(
            "{{worlds_path}}",
            &escape_html(&config.worlds.directory),
        )
        .replace(
            "{{worlds_summary}}",
            &escape_html(&worlds_summary),
        )
        .replace("{{worlds_list}}", &worlds_list)
        .replace(
            "{{csrf_token}}",
            &escape_html(&config.web.csrf_token),
        )
        .replace("{{message}}", "");

    Html(page).into_response()
}

async fn action(
    State(state): State<Arc<AppState>>,
    headers: HeaderMap,
    Form(form): Form<ActionForm>,
) -> Response {
    let config = &state.config;

    if !authorized(&headers, config) {
        return unauthorized();
    }

    if form.csrf_token != config.web.csrf_token {
        return (
            StatusCode::FORBIDDEN,
            "Invalid CSRF token",
        ).into_response();
    }

    let result = match form.action.as_str() {
        "start" => systemd::start(&config.server.systemd_unit)
            .map(|_| "Start requested".to_string()),

        "stop" => systemd::stop(&config.server.systemd_unit)
            .map(|_| "Stop requested".to_string()),

        "restart" => systemd::restart(&config.server.systemd_unit)
            .map(|_| "Restart requested".to_string()),

        "backup" => {
            let unit = config.server.systemd_unit.clone();
            let worlds = config.worlds.directory.clone();
            let backup_dir = config.backups.directory.clone();

            match tokio::task::spawn_blocking(move || {
                backups::create_consistent_backup(
                    &unit,
                    worlds,
                    backup_dir,
                )
            }).await {
                Ok(Ok(path)) => Ok(format!(
                    "Backup created: {}",
                    path.display()
                )),
                Ok(Err(error)) => Err(error),
                Err(error) => Err(anyhow::anyhow!(
                    "Backup task failed: {error}"
                )),
            }
        }

        _ => Err(anyhow::anyhow!("Unknown action")),
    };

    match result {
        Ok(message) => (
            StatusCode::OK,
            Html(format!(
                "<p>{}</p><p><a href=\"/\">Return to dashboard</a></p>",
                escape_html(&message)
            )),
        ).into_response(),

        Err(error) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Html(format!(
                "<p>Operation failed: {}</p><p><a href=\"/\">Return to dashboard</a></p>",
                escape_html(&format!("{error:#}"))
            )),
        ).into_response(),
    }
}

pub async fn serve(config: Config) -> anyhow::Result<()> {
    let address: std::net::SocketAddr = config.web.bind.parse()?;

    let state = Arc::new(AppState {
        config: Arc::new(config),
    });

    let app = Router::new()
        .route("/", get(dashboard))
        .route("/action", post(action))
        .with_state(state);

    let listener = tokio::net::TcpListener::bind(address).await?;

    tracing::info!(
        "valheimctl web listening on {}",
        listener.local_addr()?
    );

    axum::serve(listener, app).await?;
    Ok(())
}
