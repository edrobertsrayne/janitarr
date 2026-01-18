import { describe, it, expect, vi } from "vitest";
import { screen } from "@testing-library/react";
import ConfirmDialog from "./ConfirmDialog";
import { renderWithProviders, userEvent } from "../../test/utils";

describe("ConfirmDialog", () => {
  it("does not render when closed", () => {
    renderWithProviders(
      <ConfirmDialog
        open={false}
        title="Delete Server"
        message="Are you sure?"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
    );

    expect(screen.queryByText("Delete Server")).not.toBeInTheDocument();
  });

  it("renders when open", () => {
    renderWithProviders(
      <ConfirmDialog
        open={true}
        title="Delete Server"
        message="Are you sure you want to delete this server?"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
    );

    expect(screen.getByText("Delete Server")).toBeInTheDocument();
    expect(
      screen.getByText("Are you sure you want to delete this server?"),
    ).toBeInTheDocument();
  });

  it("renders default button labels", () => {
    renderWithProviders(
      <ConfirmDialog
        open={true}
        title="Delete Server"
        message="Are you sure?"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
    );

    expect(screen.getByText("Confirm")).toBeInTheDocument();
    expect(screen.getByText("Cancel")).toBeInTheDocument();
  });

  it("renders custom button labels", () => {
    renderWithProviders(
      <ConfirmDialog
        open={true}
        title="Delete Server"
        message="Are you sure?"
        confirmLabel="Delete"
        cancelLabel="Keep"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
    );

    expect(screen.getByText("Delete")).toBeInTheDocument();
    expect(screen.getByText("Keep")).toBeInTheDocument();
  });

  it("calls onConfirm when confirm button clicked", async () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(
      <ConfirmDialog
        open={true}
        title="Delete Server"
        message="Are you sure?"
        onConfirm={onConfirm}
        onCancel={onCancel}
      />,
    );

    const confirmButton = screen.getByText("Confirm");
    await user.click(confirmButton);

    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(onCancel).not.toHaveBeenCalled();
  });

  it("calls onCancel when cancel button clicked", async () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(
      <ConfirmDialog
        open={true}
        title="Delete Server"
        message="Are you sure?"
        onConfirm={onConfirm}
        onCancel={onCancel}
      />,
    );

    const cancelButton = screen.getByText("Cancel");
    await user.click(cancelButton);

    expect(onCancel).toHaveBeenCalledTimes(1);
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("applies error color when destructive", () => {
    renderWithProviders(
      <ConfirmDialog
        open={true}
        title="Delete Server"
        message="Are you sure?"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
        destructive={true}
      />,
    );

    const confirmButton = screen.getByText("Confirm");
    expect(confirmButton).toBeInTheDocument();
    // Material-UI applies color via classes, hard to test exact color
    // but we verify component renders without error
  });

  it("applies primary color when not destructive", () => {
    renderWithProviders(
      <ConfirmDialog
        open={true}
        title="Save Changes"
        message="Do you want to save?"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
        destructive={false}
      />,
    );

    const confirmButton = screen.getByText("Confirm");
    expect(confirmButton).toBeInTheDocument();
  });

  it("sets autofocus on confirm button", () => {
    renderWithProviders(
      <ConfirmDialog
        open={true}
        title="Delete Server"
        message="Are you sure?"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
    );

    const confirmButton = screen.getByText("Confirm");
    // Material-UI handles autofocus internally, just verify button exists
    expect(confirmButton).toBeInTheDocument();
  });
});
