// src/tasks.rs

use rusqlite::{params, Connection, Result};

pub fn add_task(project_name: &str, task_name: &str) -> Result<()> {
    let db_name = format!("{}.sqlite", project_name);
    let conn = Connection::open(&db_name)?;

    conn.execute(
        "INSERT INTO tasks (name, status) VALUES (?1, ?2)",
        params![task_name, "pending"],
    )?;

    println!("Task '{}' added to project '{}'", task_name, project_name);

    Ok(())
}

pub fn update_task(project_name: &str, task_id: i32, new_name: &str) -> Result<()> {
    let db_name = format!("{}.sqlite", project_name);
    let conn = Connection::open(&db_name)?;

    conn.execute(
        "UPDATE tasks SET name = ?1 WHERE id = ?2",
        params![new_name, task_id],
    )?;

    println!("Task with ID {} updated to '{}'", task_id, new_name);

    Ok(())
}

pub fn move_status_forward(project_name: &str, task_id: i32) -> Result<()> {
    let db_name = format!("{}.sqlite", project_name);
    let conn = Connection::open(&db_name)?;

    let mut stmt = conn.prepare("SELECT status FROM tasks WHERE id = ?1")?;
    let mut rows = stmt.query(params![task_id])?;

    let next_status = if let Some(row) = rows.next()? {
        let current_status: String = row.get(0)?;

        match current_status.as_str() {
            "pending" => "in_progress",
            "in_progress" => "completed",
            "completed" => "completed",
            _ => {
                eprintln!("Invalid task status: {}", current_status);
                return Ok(());
            }
        }
    } else {
        eprintln!("Task with ID {} not found", task_id);
        return Ok(());
    };

    conn.execute(
        "UPDATE tasks SET status = ?1 WHERE id = ?2",
        params![next_status, task_id],
    )?;

    println!("Task with ID {} status moved to '{}'", task_id, next_status);

    Ok(())
}
