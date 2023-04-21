// src/main.rs

use clap::Parser;
use std::env;
use std::process;

mod init;

fn main() {
    let matches = command!() // requires `cargo` feature
        .arg(arg!([name] "Optional name to operate on"))
        .arg(
            arg!(
                -c --config <FILE> "Sets a custom config file"
            )
            // We don't have syntax yet for optional options, so manually calling `required`
            .required(false)
            .value_parser(value_parser!(PathBuf)),
        )
        .arg(arg!(
            -d --debug ... "Turn debugging information on"
        ))
        .subcommand(
            Command::new("test")
                .about("does testing things")
                .arg(arg!(-l --list "lists test values").action(ArgAction::SetTrue)),
        )
        .get_matches();
    let args: Vec<String> = env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: plannet <command> [options]");
        process::exit(1);
    }

    let command = &args[1];

    match command.as_str() {
        "init" => {
            if args.len() < 3 {
                eprintln!("Usage: plannet init <project_name>");
                process::exit(1);
            }
            let project_name = &args[2];
            if let Err(e) = init::init(project_name) {
                eprintln!("Error initializing project: {}", e);
                process::exit(1);
            }
        }
        "add" => {
            if args.len() < 4 {
                eprintln!("Usage: plannet add <project_name> <task_name>");
                process::exit(1);
            }
        }
        _ => {
            eprintln!("Unknown command: {}", command);
            process::exit(1);
        }
    }
}
