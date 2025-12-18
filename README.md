# Breakfront Planner

Task planning tool that displays tasks across three time perspectives: day, week, and month.

## Features

- Create, update, delete tasks
- Assign tasks to day, week, or month
- View tasks grouped by time period
- JWT authentication

## Period Resolution Rules

- Day task → appears in that day, its week, and its month
- Week task → appears in that week and its month(s). If week spans two months — appears in both
- Month task → appears in that month

## Tech Stack

- **Backend:** Go
- **Database:** PostgreSQL
- **Auth:** JWT + bcrypt
- **Deploy:** Docker Compose


## Status

🚧 Work in progress — Auth Service in development