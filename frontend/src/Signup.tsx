import React, { useState } from "react";
import { Form, Input, Join, Button } from "react-daisyui";
import { useNavigate } from "@tanstack/react-router";

function Signup() {
  const [name, setName] = useState<{ name: string }>("");
  const [email, setEmail] = useState<{ email: string }>("");
  const [password, setPassword] = useState<{ password: string }>("");
  const [partyID, setPartyID] = useState<{ partyID: string }>("");

  const navigate = useNavigate({ from: "/login" });

  function handleOnClick(e: ChangeEvent<HTMLInputElement>) {
    e.preventDefault();
    const signUp = async () => {
      try {
        const response = await fetch("http://localhost:8080/signup", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            name: name,
            login: email,
            password: password,
            partyID: partyID,
          }),
        });
        console.log(response.status);
        if (response.ok) {
          navigate({ to: "/login" });
        }
      } catch (error) {
        console.error(error);
      }
    };
    signUp();
  }

  return (
    <div>
      <h3>Signup</h3>
      <Form>
        <Join>
          <div className="columns-2">
            <Form.Label title="Name" />
          </div>
          <div className="columns-1">
            <Input
              type="text"
              value={name}
              placeholder="Name"
              className="input-bordered"
              onChange={(e: ChangeEvent<HTMLInputElement>) =>
                setName(e.target.value)
              }
            />
          </div>
        </Join>
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
        <Join>
          <Form.Label title="Party ID" />
          <Input
            type="text"
            placeholder="Party ID"
            className="input-bordered"
            onChange={(e) => setPartyID(e.target.value)}
            value={partyID}
          />
        </Join>
        <div className="columns-2">
          <Button tag="button" onClick={handleOnClick}>
            Sign Up!
          </Button>
        </div>
      </Form>
    </div>
  );
}

export default Signup;
