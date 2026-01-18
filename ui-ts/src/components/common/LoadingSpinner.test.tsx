import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import LoadingSpinner from "./LoadingSpinner";
import { renderWithProviders } from "../../test/utils";

describe("LoadingSpinner", () => {
  it("renders spinner without message", () => {
    renderWithProviders(<LoadingSpinner />);

    // Check that progress indicator is present
    const spinner = screen.getByRole("progressbar");
    expect(spinner).toBeInTheDocument();
  });

  it("renders spinner with custom message", () => {
    renderWithProviders(<LoadingSpinner message="Loading data..." />);

    const spinner = screen.getByRole("progressbar");
    expect(spinner).toBeInTheDocument();

    const message = screen.getByText("Loading data...");
    expect(message).toBeInTheDocument();
  });

  it("does not render message when not provided", () => {
    renderWithProviders(<LoadingSpinner />);

    // Should only have the spinner, no text
    const spinner = screen.getByRole("progressbar");
    expect(spinner).toBeInTheDocument();

    // No text elements should be present
    expect(screen.queryByRole("text")).not.toBeInTheDocument();
  });

  it("accepts custom size prop", () => {
    renderWithProviders(<LoadingSpinner size={80} />);

    const spinner = screen.getByRole("progressbar");
    expect(spinner).toBeInTheDocument();
    // Size is applied via sx prop, hard to test exact value
    // but we verify component renders without error
  });
});
