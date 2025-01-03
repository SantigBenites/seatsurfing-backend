import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

export default class SpaceAttribute extends Entity {
    label: string;
    type: number;
    spaceApplicable: boolean;
    locationApplicable: boolean;

    constructor() {
        super();
        this.label = "";
        this.type = 1;
        this.spaceApplicable = false;
        this.locationApplicable = false;
    }

    serialize(): Object {
        return Object.assign(super.serialize(), {
            "label": this.label,
            "type": this.type,
            "spaceApplicable": this.spaceApplicable,
            "locationApplicable": this.locationApplicable,
        });
    }

    deserialize(input: any): void {
        super.deserialize(input);
        this.label = input.label;
        this.type = input.type;
        this.spaceApplicable = input.spaceApplicable;
        this.locationApplicable = input.locationApplicable;
    }

    getBackendUrl(): string {
        return "/space-attribute/";
    }

    async save(): Promise<SpaceAttribute> {
        return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
    }

    async delete(): Promise<void> {
        return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
    }

    static async get(id: string): Promise<SpaceAttribute> {
        return Ajax.get("/space-attribute/" + id).then(result => {
            let e: SpaceAttribute = new SpaceAttribute();
            e.deserialize(result.json);
            return e;
        });
    }

    static async list(): Promise<SpaceAttribute[]> {
        return Ajax.get("/space-attribute/").then(result => {
            let list: SpaceAttribute[] = [];
            (result.json as []).forEach(item => {
                let e: SpaceAttribute = new SpaceAttribute();
                e.deserialize(item);
                list.push(e);
            });
            return list;
        });
    }
}
