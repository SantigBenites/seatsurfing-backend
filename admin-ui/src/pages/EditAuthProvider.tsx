import React from 'react';
import FullLayout from '../components/FullLayout';
import { Form, Col, Row, Button, Alert } from 'react-bootstrap';
import { ChevronLeft as IconBack, Save as IconSave, Trash2 as IconDelete } from 'react-feather';
import { Link, RouteChildrenProps, Redirect } from 'react-router-dom';
import Loading from '../components/Loading';
import { AuthProvider } from 'flexspace-commons';

interface State {
  loading: boolean
  submitting: boolean
  saved: boolean
  goBack: boolean
  name: string
  providerType: number;
  authUrl: string;
  tokenUrl: string;
  authStyle: number;
  scopes: string;
  userInfoUrl: string;
  userInfoEmailField: string;
  clientId: string;
  clientSecret: string;
}

interface Props {
  id: string
}

export default class EditAuthProvider extends React.Component<RouteChildrenProps<Props>, State> {
  entity: AuthProvider = new AuthProvider();

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      goBack: false,
      name: "",
      providerType: 0,
      authUrl: "",
      tokenUrl: "",
      authStyle: 0,
      scopes: "",
      userInfoUrl: "",
      userInfoEmailField: "",
      clientId: "",
      clientSecret: ""
    };
  }

  componentDidMount = () => {
    this.loadData();
  }

  loadData = (id?: string) => {
    if (!id) {
      id = this.props.match?.params.id;
    }
    if (id) {
      AuthProvider.get(id).then(authProvider => {
        this.entity = authProvider;
        this.setState({
          name: authProvider.name,
          providerType: authProvider.providerType,
          authUrl: authProvider.authUrl,
          tokenUrl: authProvider.tokenUrl,
          authStyle: authProvider.authStyle,
          scopes: authProvider.scopes,
          userInfoUrl: authProvider.userInfoUrl,
          userInfoEmailField: authProvider.userInfoEmailField,
          clientId: authProvider.clientId,
          clientSecret: authProvider.clientSecret,
          loading: false
        });
      });
    } else {
      this.setState({
        loading: false
      });
    }
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.entity.name = this.state.name;
    this.entity.providerType = this.state.providerType;
    this.entity.authUrl = this.state.authUrl;
    this.entity.tokenUrl = this.state.tokenUrl;
    this.entity.authStyle = this.state.authStyle;
    this.entity.scopes = this.state.scopes;
    this.entity.userInfoUrl = this.state.userInfoUrl;
    this.entity.userInfoEmailField = this.state.userInfoEmailField;
    this.entity.clientId = this.state.clientId;
    this.entity.clientSecret = this.state.clientSecret;
    this.entity.save().then(() => {
      this.props.history.push("/auth-providers/" + this.entity.id);
      this.setState({
        saved: true
      });
    });
  }

  deleteItem = () => {
    if (window.confirm("Provider löschen? Logins über diesen Provider sind dann nicht mehr möglich!")) {
      this.entity.delete().then(() => {
        this.setState({ goBack: true });
      });
    }
  }

  templateGoogle = () => {
    this.setState({
      name: "Google",
      providerType: 1,
      authUrl: "https://accounts.google.com/o/oauth2/auth",
      tokenUrl: "https://oauth2.googleapis.com/token",
      authStyle: 1,
      scopes: "https://www.googleapis.com/auth/userinfo.email",
      userInfoUrl: "https://www.googleapis.com/oauth2/v3/userinfo",
      userInfoEmailField: "email"
    });
  }

  render() {
    if (this.state.goBack) {
      return <Redirect to={`/settings`} />
    }

    let backButton = <Link to="/settings" className="btn btn-sm btn-outline-secondary"><IconBack className="feather" /> Zurück</Link>;
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout headline="Auth Provider bearbeiten" buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success">Eintrag wurde aktualisiert.</Alert>
    }
    
    let urlInfo = <></>;
    let buttonDelete = <Button className="btn-sm" variant="outline-secondary" onClick={this.deleteItem}><IconDelete className="feather" /> Löschen</Button>;
    let buttonSave = <Button className="btn-sm" variant="outline-secondary" type="submit" form="form"><IconSave className="feather" /> Speichern</Button>;
    if (this.entity.id) {
      buttons = <>{backButton} {buttonDelete} {buttonSave}</>;
      urlInfo = (
        <Form.Group as={Row}>
          <Form.Label column sm="2">Callback URL</Form.Label>
          <Col sm="4">
            <Form.Control plaintext={true} readOnly={true} onClick={(e: any) => e.target.select()} defaultValue={`/auth/${this.entity.id}/callback`} />
          </Col>
        </Form.Group>
      );
    } else {
      buttons = <>{backButton} {buttonSave}</>;
    }
    return (
      <FullLayout headline="Auth Provider bearbeiten" buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">Name</Form.Label>
            <Col sm="4">
              <Form.Control type="text" placeholder="Name" value={this.state.name} onChange={(e: any) => this.setState({ name: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Typ</Form.Label>
            <Col sm="4">
              <Form.Control as="select" value={this.state.providerType} onChange={(e: any) => this.setState({ providerType: parseInt(e.target.value) })} required={true}>
                <option value="0">(bitte auswählen)</option>
                <option value="1">OAuth 2</option>
              </Form.Control>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Auth URL</Form.Label>
            <Col sm="4">
              <Form.Control type="url" placeholder="https://..." value={this.state.authUrl} onChange={(e: any) => this.setState({ authUrl: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Token URL</Form.Label>
            <Col sm="4">
              <Form.Control type="url" placeholder="https://..." value={this.state.tokenUrl} onChange={(e: any) => this.setState({ tokenUrl: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Auth Style</Form.Label>
            <Col sm="4">
              <Form.Control as="select" value={this.state.authStyle} onChange={(e: any) => this.setState({ authStyle: parseInt(e.target.value) })} required={true}>
                <option value="0">Automatisch</option>
                <option value="1">Parameter (HTTP POST body)</option>
                <option value="2">Header (HTTP Basic Authorization)</option>
              </Form.Control>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Scopes</Form.Label>
            <Col sm="4">
              <Form.Control type="text" placeholder="scope1,scope2,..." value={this.state.scopes} onChange={(e: any) => this.setState({ scopes: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Client ID</Form.Label>
            <Col sm="4">
              <Form.Control type="text" placeholder="Client ID" value={this.state.clientId} onChange={(e: any) => this.setState({ clientId: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Client Secret</Form.Label>
            <Col sm="4">
              <Form.Control type="text" placeholder="Client Secret" value={this.state.clientSecret} onChange={(e: any) => this.setState({ clientSecret: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Userinfo URL</Form.Label>
            <Col sm="4">
              <Form.Control type="url" placeholder="https://..." value={this.state.userInfoUrl} onChange={(e: any) => this.setState({ userInfoUrl: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">Userinfo E-Mail-Feld</Form.Label>
            <Col sm="4">
              <Form.Control type="text" placeholder="email" value={this.state.userInfoEmailField} onChange={(e: any) => this.setState({ userInfoEmailField: e.target.value })} required={true} />
            </Col>
          </Form.Group>
          {urlInfo}
          <Form.Group as={Row}>
            <Form.Label column sm="2">Vorlagen</Form.Label>
            <Col sm="4">
              <Button variant="outline-secondary" onClick={this.templateGoogle}>Google</Button>
            </Col>
          </Form.Group>
        </Form>
      </FullLayout>
    );
  }
}