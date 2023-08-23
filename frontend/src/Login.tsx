import React, { useState, useContext } from "react";
import { Form, Input, Join, Button } from "react-daisyui";
import { useNavigate } from "@tanstack/react-router";
import { AuthenticatedDispatchContext } from "./Context.tsx";

function Login() {
  const [email, setEmail] = useState<{ email: string }>("");
  const [password, setPassword] = useState<{ password: string }>("");
  const authenticatedDispatch = useContext(AuthenticatedDispatchContext);

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
          authenticatedDispatch(true);
          localStorage.setItem("moviesauth", "true");
          navigate({ to: "/" });
        }
      } catch (error) {
        console.error(error);
      }
    };
    login();
  }

  return (
    <div className="flex justify-center">
      <div className="prose">
        <h1>Login</h1>
        <Form className="space-y-5">
          <Join className="place-content-center space-x-5">
            <Form.Label title="Email" />
            <Input
              type="text"
              value={email}
              placeholder="Email"
              className="input-bordered"
              onChange={(e: ChangeEvent<HTMLInputElement>) =>
                setEmail(e.target.value)
              }
            />
          </Join>
          <Join className="place-content-center space-x-5">
            <Form.Label title="Password" />
            <Input
              type="password"
              placeholder="Password"
              className="input-bordered"
              onChange={(e: ChangeEvent<HTMLInputElement>) =>
                setPassword(e.target.value)
              }
              value={password}
            />
          </Join>
          <div className="place-content-center space-x-5">
            <Button tag="button" color="primary" onClick={handleOnClick}>
              Click to Login
            </Button>
          </div>
        </Form>
      </div>
    </div>
  );
}

export default Login;
