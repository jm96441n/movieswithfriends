import React, { useState } from "react";
import { Form, Input, Join, Button } from "react-daisyui";
import { useNavigate } from "@tanstack/react-router";

function Login() {
  const [email, setEmail] = useState<{ email: string }>("");
  const [password, setPassword] = useState<{ password: string }>("");

  const navigate = useNavigate({ from: "/" });

  function handleOnClick(e: ChangeEvent<HTMLInputElement>) {
    e.preventDefault();
    const login = async () => {
      try {
        const response = await fetch("http://localhost:8080/login", {
          method: "POST",
          mode: "cors",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            login: email,
            password: password,
          }),
        });
        if (response.ok) {
          navigate({ to: "/" });
        }
      } catch (error) {
        console.error(error);
      }
    };
    login();
  }

  return (
    <div>
      <h3>Login</h3>
      <Form>
        <Join>
          <div className="columns-2">
            <Form.Label title="Email" />
          </div>
          <div className="columns-1">
            <Input
              type="text"
              value={email}
              placeholder="Email"
              className="input-bordered"
              onChange={(e: ChangeEvent<HTMLInputElement>) =>
                setEmail(e.target.value)
              }
            />
          </div>
        </Join>
        <Join>
          <div className="columns-2">
            <Form.Label title="Password" />
          </div>
          <div className="columns-1">
            <Input
              type="password"
              placeholder="Password"
              className="input-bordered"
              onChange={(e: ChangeEvent<HTMLInputElement>) =>
                setPassword(e.target.value)
              }
              value={password}
            />
          </div>
        </Join>
        <div className="columns-2">
          <Button tag="button" onClick={handleOnClick}>
            Click to Login
          </Button>
        </div>
      </Form>
    </div>
  );
}

export default Login;
