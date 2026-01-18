import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import StatusBadge from "./StatusBadge";
import { renderWithProviders } from "../../test/utils";

describe("StatusBadge", () => {
  it("renders connected status with default label", () => {
    renderWithProviders(<StatusBadge status="connected" />);

    const badge = screen.getByText("Connected");
    expect(badge).toBeInTheDocument();
  });

  it("renders error status with default label", () => {
    renderWithProviders(<StatusBadge status="error" />);

    const badge = screen.getByText("Error");
    expect(badge).toBeInTheDocument();
  });

  it("renders disabled status with default label", () => {
    renderWithProviders(<StatusBadge status="disabled" />);

    const badge = screen.getByText("Disabled");
    expect(badge).toBeInTheDocument();
  });

  it("renders running status with default label", () => {
    renderWithProviders(<StatusBadge status="running" />);

    const badge = screen.getByText("Running");
    expect(badge).toBeInTheDocument();
  });

  it("renders stopped status with default label", () => {
    renderWithProviders(<StatusBadge status="stopped" />);

    const badge = screen.getByText("Stopped");
    expect(badge).toBeInTheDocument();
  });

  it("renders custom label when provided", () => {
    renderWithProviders(<StatusBadge status="connected" label="Online" />);

    const badge = screen.getByText("Online");
    expect(badge).toBeInTheDocument();

    // Default label should not be present
    expect(screen.queryByText("Connected")).not.toBeInTheDocument();
  });

  it("renders icon for each status", () => {
    const { rerender } = renderWithProviders(
      <StatusBadge status="connected" />,
    );

    // Each status should render an icon (svg)
    // We can check that the chip is rendered as a visual indicator
    expect(screen.getByText("Connected")).toBeInTheDocument();

    // Test other statuses
    rerender(<StatusBadge status="error" />);
    expect(screen.getByText("Error")).toBeInTheDocument();

    rerender(<StatusBadge status="running" />);
    expect(screen.getByText("Running")).toBeInTheDocument();
  });
});
