# Current Issues

- Clicking buttons in the web interface does not open dialogs for add / edit server
- Logs in the web interface do not match logs in the CLI
- The Run Now button on the dashboard appears odd, is there a missing icon that's not showing?
- Theme choser is still on the Settings page, despite being removed and replaced with a light/dark toggle. Needs to be removed.
- The servers list on the dashboard has a field for URL that doesn't show anything.
- Clicking "test" on the Servers card gives a "connection failed" error despite the server working as expected.
- The edit server button does nothing
- The server delete button opens a browser modal rather than a DaisyUI modal
- Starting the CLI tool in dev mode should use a different port than the production mode's 3434
- If a dev or production server is started, it should check if its preferred port is free and if not fall back to the next available port
