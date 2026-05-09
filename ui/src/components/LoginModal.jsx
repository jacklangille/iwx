import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { login, signup } from "../lib/api";
import { useAuth } from "../lib/auth";

export function LoginModal({ open, mode = "login", onClose }) {
  const { login: setSession } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [localError, setLocalError] = useState("");
  const signupMode = mode === "signup";

  const mutation = useMutation({
    mutationFn: ({ nextUsername, nextPassword }) =>
      signupMode ? signup(nextUsername, nextPassword) : login(nextUsername, nextPassword),
    onSuccess(payload) {
      setSession({
        accessToken: payload.access_token,
        userId: payload.user_id,
        subject: payload.subject,
        expiresAt: payload.expires_at,
      });
      setUsername("");
      setPassword("");
      setConfirmPassword("");
      setLocalError("");
      onClose();
    },
  });

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="xs">
      <DialogTitle>{signupMode ? "Sign up" : "Log in"}</DialogTitle>
      <form
        onSubmit={(event) => {
          event.preventDefault();
          setLocalError("");
          if (signupMode && password !== confirmPassword) {
            setLocalError("Passwords must match");
            return;
          }
          mutation.mutate({ nextUsername: username, nextPassword: password });
        }}
      >
        <DialogContent dividers>
          <Stack spacing={2}>
            <TextField
              label="Username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              fullWidth
              autoFocus
            />

            <TextField
              label="Password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete={signupMode ? "new-password" : "current-password"}
              fullWidth
            />

            {signupMode ? (
              <TextField
                label="Confirm password"
                type="password"
                value={confirmPassword}
                onChange={(event) => setConfirmPassword(event.target.value)}
                autoComplete="new-password"
                fullWidth
              />
            ) : null}

            {localError ? <Alert severity="error">{localError}</Alert> : null}

            {mutation.error ? (
              <Alert severity="error">
                {mutation.error.payload?.error || mutation.error.message}
              </Alert>
            ) : null}

            {signupMode ? (
              <Typography variant="body2" color="text.secondary">
                New accounts are created immediately and signed in automatically.
              </Typography>
            ) : null}
          </Stack>
        </DialogContent>

        <DialogActions>
          <Button onClick={onClose} color="inherit">
            Close
          </Button>
          <Button type="submit" variant="contained" disabled={mutation.isPending}>
            {mutation.isPending
              ? signupMode
                ? "Creating account..."
                : "Logging in..."
              : signupMode
                ? "Create account"
                : "Log in"}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}
