import React, { useState } from "react";
import { Form, Input, Join, Button } from "react-daisyui";

function Signup() {
  const [email, setEmail] = useState<{ email: string }>("");
  const [password, setPassword] = useState<{ password: string }>("");
  const [partyID, setPartyID] = useState<{ partyID: string }>("");

  function handleOnClick(e: ChangeEvent<HTMLInputElement>) {
    e.preventDefault()
    const fetchData = async() => {
      try {
        const response = await fetch("http://localhost:8080/signup", {
          method: "POST",
          mode: 'no-cors',
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            login: email,
            password: password,
            partyID: partyID,
          })
        });
        const body = await response.json()
        const status = response.status
        alert(`${status} ${body}`)
      } catch (error) {
        alert(error);
      }
    }
    fetchData();
  }

  return (
    <div>
      <h3>Signup</h3>
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
