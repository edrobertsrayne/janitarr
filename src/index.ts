#!/usr/bin/env bun

/**
 * Janitarr - Automation tool for Radarr and Sonarr media servers
 *
 * Entry point for the CLI application.
 */

import { createProgram } from "./cli/commands";

const program = createProgram();
program.parse(process.argv);
