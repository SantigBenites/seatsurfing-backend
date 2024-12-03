import React from 'react';
import {Ajax, AjaxError, Booking, Formatting} from 'flexspace-commons';
import Loading from '../components/Loading';
import { Button, Form, ListGroup, Modal } from 'react-bootstrap';
import { LogIn as IconEnter, LogOut as IconLeave, MapPin as IconLocation } from 'react-feather';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import NavBar from '@/components/NavBar';
import withReadyRouter from '@/components/withReadyRouter';
import RuntimeConfig from '@/components/RuntimeConfig';
import ErrorText from '@/types/ErrorText';

interface State {
  loading: boolean
  selectedItem: Booking | null
}

interface Props extends WithTranslation {
  router: NextRouter
}

class Bookings extends React.Component<Props, State> {
  data: Booking[];

  constructor(props: any) {
    super(props);
    this.data = [];
    this.state = {
      loading: true,
      selectedItem: null
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    this.loadData();
  }

  loadData = () => {
    Booking.list().then(list => {
      this.data = list;
      this.setState({ loading: false });
    });
  }

  onItemPress = (item: Booking) => {
    this.setState({ selectedItem: item });
  }

  cancelBooking = (item: Booking | null) => {
    this.setState({
      loading: true
    });
    this.state.selectedItem?.delete().then(() => {
      this.setState({
        selectedItem: null,
      }, this.loadData);
    }, (reason: any) => {
      if (reason instanceof AjaxError && reason.httpStatusCode === 403) {
          window.alert(ErrorText.getTextForAppCode(reason.appErrorCode, this.props.t));
        } else {
          window.alert(this.props.t("errorDeleteBooking"));
        }
      this.setState({
        selectedItem: null,
      }, this.loadData);
    });
  }

  renderItem = (item: Booking) => {
    let formatter = Formatting.getFormatter();
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      formatter = Formatting.getFormatterNoTime();
    }
    return (
      <ListGroup.Item key={item.id} action={true} onClick={(e) => { e.preventDefault(); this.onItemPress(item); }}>
        <h5>{Formatting.getDateOffsetText(item.enter, item.leave)}</h5>
        <p>
          <IconLocation className="feather" />&nbsp;{item.space.location.name}, {item.space.name}<br />
          <IconEnter className="feather" />&nbsp;{formatter.format(item.enter)}<br />
          <IconLeave className="feather" />&nbsp;{formatter.format(item.leave)}
        </p>
      </ListGroup.Item>
    );
  }

  groupBookingsByDate(bookings: Booking[]): Record<string, Booking[]> {
    return bookings.reduce((groups: Record<string, Booking[]>, booking: Booking) => {
      const date = new Date(booking.enter).toLocaleDateString();
      if (!groups[date]) {
        groups[date] = [];
      }
      groups[date].push(booking);
      return groups;
    }, {});
  }
  
  

  groupBookingsByMonth(bookings: Booking[]): Record<string, Booking[]> {
    return bookings.reduce((groups: Record<string, Booking[]>, booking: Booking) => {
      const date = new Date(booking.enter);
      const monthYear = `${date.toLocaleString('default', { month: 'long' })} ${date.getFullYear()}`;
      if (!groups[monthYear]) {
        groups[monthYear] = [];
      }
      groups[monthYear].push(booking);
      return groups;
    }, {});
  }
  
  render() {
    if (this.state.loading) {
      return <Loading />;
    }
    if (this.data.length === 0) {
      return (
        <>
          <NavBar />
          <div className="container-signin">
            <Form className="form-signin">
              <p>{this.props.t("noBookings123")}</p>
            </Form>
          </div>
        </>
      );
    }
  
    const groupedBookings = this.groupBookingsByMonth(this.data);
  
    return (
      <>
        <NavBar />
        <div className="container-signin">
          <Form className="form-signin" style={{ width: "100%" }}>
            <div className="monthly-bookings-container">
              {Object.entries(groupedBookings).map(([month, bookings]) => (
                <div key={month} className="month-group">
                  <h4>{month}</h4>
                  <ListGroup horizontal>
                    {bookings.map((item) => this.renderItem(item))}
                  </ListGroup>
                </div>
              ))}
            </div>
          </Form>
        </div>
        <Modal
          show={this.state.selectedItem != null}
          onHide={() => this.setState({ selectedItem: null })}
        >
          <Modal.Header closeButton>
            <Modal.Title>{this.props.t("cancelBooking")}</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <p>
              {this.props.t("confirmCancelBooking", {
                enter: new Intl.DateTimeFormat('default', {
                  dateStyle: 'full',
                  timeStyle: 'short',
                }).format(this.state.selectedItem?.enter),
              })}
            </p>
          </Modal.Body>
          <Modal.Footer>
            <Button
              variant="secondary"
              onClick={() => this.setState({ selectedItem: null })}
            >
              {this.props.t("back")}
            </Button>
            <Button
              variant="danger"
              onClick={() => this.cancelBooking(this.state.selectedItem)}
            >
              {this.props.t("cancelBooking")}
            </Button>
          </Modal.Footer>
        </Modal>
      </>
    );
  }
  
  

}

export default withTranslation()(withReadyRouter(Bookings as any));
