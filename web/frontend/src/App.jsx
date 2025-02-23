import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';

function App() {
  const [count, setCount] = useState(0)

  var url = "ws://" + window.location.host + "/ws";
  var ws = new WebSocket(url);
  var mypid = localStorage.getItem("pid");
  var mywid = "";

  var renderQueue = [];

  return (
    <>
      <h1>Liar's Fortress</h1>

      <Stack spacing={2} alignItems="center">
        <TextField id="outlined-basic" label="Lobby ID" variant="outlined"/>
        <Stack spacing={2} direction="row" alignItems="justify">
          <Button variant="contained">Join Lobby</Button>
          <Button variant="contained">Create Lobby</Button>
        </Stack>
      </Stack>

    </>
  )
}

export default App