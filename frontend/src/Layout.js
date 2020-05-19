import React from 'react';
import Overview from './Overview';
import Navbar from 'react-bootstrap/Navbar';
import Nav from 'react-bootstrap/Nav';
import NavDropdown from 'react-bootstrap/NavDropdown';
import Container from 'react-bootstrap/Container';
import { Link } from 'react-router-dom';
import 'bootstrap/dist/css/bootstrap.min.css';
import './Layout.css';
import Details from './Details';

function Layout(props) {
    let content = <Overview currentHost={props.currentHost} hosts={props.hosts} />;
    if (props.pageNotFound) {
        content = <div className="text-center"><h3>Page Not Found</h3></div>
    }
    if (props.hostId) {
        content = <Details hostId={props.hostId} />
    }

    return (
        <div>
            <Navbar bg="dark" expand="md" variant="dark">
                <Navbar.Toggle aria-controls="navbarSupportedContent" />
                <Navbar.Collapse id="navbarSupportedContent">
                    <Nav className="mr-auto">
                        <NavDropdown id="navbarDropdown" title={props.currentHost.hostname} active>
                            {props.hosts.map((item, key) => {
                                return (
                                    <NavDropdown.Item
                                        key={key}
                                        href={"http://" + item.publicIp}
                                    >
                                        {item.hostname}
                                    </NavDropdown.Item>
                                );
                            })}
                        </NavDropdown>

                        <Link className="nav-link" to="/">
                            Home
                        </Link>
                    </Nav>
                </Navbar.Collapse>
            </Navbar>

            <Container fluid className="main-container">
                <Container fluid className="content-container">{content}</Container>
            </Container>
        </div>
    );
}

export default Layout;
