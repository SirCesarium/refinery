//! Error management for Refinery-RS.
//!
//! Provides a centralized error type using `thiserror` and `miette`
//! for beautiful diagnostic reporting.
#![allow(dead_code)]

use miette::Diagnostic;
use std::io::Error as StdIoError;
use std::result::Result as StdResult;
use thiserror::Error;

/// Core error type for the refinery CLI.
///
/// This enum encompasses all possible errors that can occur within the Refinery-RS
/// application, providing structured information for diagnostics and error handling.
#[derive(Debug, Error, Diagnostic)]
pub enum RefineryError {
    /// IO-related failures when managing workflow files.
    ///
    /// This variant wraps `std::io::Error` and provides context for file operations.
    #[error("Failed to perform IO operation: {0}")]
    #[diagnostic(
        code(refinery::io_error),
        help("Check file permissions or if the directory exists.")
    )]
    Io(#[from] StdIoError),

    /// Errors occurring during the interactive prompt session.
    ///
    /// This variant wraps `inquire::InquireError` for user input related issues.
    #[error("Failed to process interactive prompt: {0}")]
    #[diagnostic(code(refinery::prompt_error))]
    Prompt(#[from] inquire::InquireError),

    /// Error thrown when a workflow file already exists and force is not used.
    ///
    /// Contains the name of the file that already exists.
    #[error("Workflow file '{0}' already exists.")]
    #[diagnostic(
        code(refinery::file_exists),
        help("Use the --force flag to overwrite existing workflow files.")
    )]
    FileExists(String),

    /// Serialization/Deserialization errors for YAML.
    ///
    /// Wraps `serde_yaml::Error` for issues with YAML configuration files.
    #[error("Failed to process YAML configuration: {0}")]
    #[diagnostic(code(refinery::yaml_error))]
    Yaml(#[from] serde_yaml::Error),

    /// Network related errors for updates or external API calls.
    ///
    /// Wraps `reqwest::Error` for problems during network requests.
    #[error("Network operation failed: {0}")]
    #[diagnostic(code(refinery::network_error))]
    Network(#[from] reqwest::Error),
}

/// Type alias for Refinery results.
///
/// This is a convenience alias for `std::result::Result<T, RefineryError>`,
/// simplifying function signatures that return a `RefineryError`.
pub type Result<T> = StdResult<T, RefineryError>;
