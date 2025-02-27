import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';

function UsernameRegister({ws, setUsername, setPage, username}) {

    function register() {
        var username = document.getElementById("username").value;
        localStorage.setItem("LFusername", username);
        ws.send(`name ${username}`);
        setUsername(username);
        setPage("lobbyjoin");
    }

    const handleChange = (event) => {
        setUsername(event.target.value);
    }

    return (
        <>
            <h1>Liar's Fortress</h1>

            <Stack spacing={2} alignItems="center">
                <TextField id="username" label="Username" variant="outlined" value={username} onChange={handleChange}/>
                <Stack spacing={2} direction="row" alignItems="justify">
                    <Button variant="contained" onClick={register}>Register</Button>
                </Stack>
            </Stack>

        </>
    )
}

export default UsernameRegister