interface Update {
    selector: string;
    text: { [key: string]: string };
    integer: { [key: string]: number };
    boolean: { [key: string]: boolean };
    attribute: { [key: string]: string };
    style: { [key: string]: string };
    remove: boolean;
}