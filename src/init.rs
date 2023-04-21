use std::fs::{self, File};
use std::io::prelude::*;
use std::path::Path;

use rusqlite::{Connection, Result};

pub fn init(project_name: &str) -> Result<(), Box<dyn std::error::Error>> {
    create_project_folder(project_name)?;
    create_plan_file(project_name)?;
    create_database(project_name)?;
    create_plannetrc_file(project_name)?;
    Ok(())
}

fn create_project_folder(project_name: &str) -> std::io::Result<()> {
    let project_folder = Path::new(project_name);
    fs::create_dir(&project_folder)?;
    println!("Created project folder: {:?}", project_folder);
    Ok(())
}

fn create_plan_file(project_name: &str) -> std::io::Result<()> {
    let file_name = format!("{}/{}.plan", project_name, project_name);
    let mut file = File::create(&file_name)?;
    writeln!(file, "Project Plan: {}", project_name)?;
    println!("Created plan file: {}", file_name);
    Ok(())
}

fn create_database(project_name: &str) -> Result<()> {
    let db_name = format!("{}/{}.sqlite", project_name, project_name);
    let conn = Connection::open(&db_name)?;

    conn.execute(
        "CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            status TEXT NOT NULL
            description BLOB
        )",
        [],
    )?;

    println!("Created SQLite database: {}", db_name);
    Ok(())
}

fn create_plannetrc_file(project_name: &str) -> std::io::Result<()> {
    let file_name = format!("{}/.plannetrc", project_name);
    let mut file = File::create(&file_name)?;
    writeln!(file, "project_name = \"{}\"", project_name)?;
    println!("Created .plannetrc file: {}", file_name);
    Ok(())
}
